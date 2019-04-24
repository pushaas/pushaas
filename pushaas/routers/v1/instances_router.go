package v1

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rafaeleyng/pushaas/pushaas/services"

	"github.com/rafaeleyng/pushaas/pushaas/routers"

	"github.com/rafaeleyng/pushaas/pushaas/business"
)

type (
	instanceRouter struct {
		instanceService services.InstanceService
		logger          *log.Logger
	}
)

func (r *instanceRouter) postInstance(c *gin.Context) {
	// sess, err := session.NewSession()
	// if err != nil {
	// 	fmt.Errorf("err", err)
	// 	return
	// }

	//	// svc := ecs.New(sess)
	//
	//	// CreateService
	//	// RegisterTaskDefinition
	//	// RunTask | StartTask

	// svc := ec2.New(sess)

	// // Specify the details of the instance that you want to create.
	// runResult, err := svc.RunInstances(&ec2.RunInstancesInput{
	// 	// An Amazon Linux AMI ID for t2.micro instances in the us-west-2 region
	// 	ImageId:      aws.String("ami-0de53d8956e8dcf80"),
	// 	InstanceType: aws.String("t2.micro"),
	// 	MinCount:     aws.Int64(1),
	// 	MaxCount:     aws.Int64(1),
	// })

	// if err != nil {
	// 	fmt.Println("Could not create instance", err)
	// 	return
	// }

	// fmt.Println("Created instance", *runResult.Instances[0].InstanceId)

	// // Add tags to the created instance
	// _, errtag := svc.CreateTags(&ec2.CreateTagsInput{
	// 	Resources: []*string{runResult.Instances[0].InstanceId},
	// 	Tags: []*ec2.Tag{
	// 		{
	// 			Key:   aws.String("Name"),
	// 			Value: aws.String("MyFirstInstance"),
	// 		},
	// 	},
	// })
	// if errtag != nil {
	// 	log.Println("Could not create tags for instance", runResult.Instances[0].InstanceId, errtag)
	// 	return
	// }

	// fmt.Println("Successfully tagged instance")

	toCreate := &services.Instance{
		Description: "an instance of a push service",
	}

	instance, err := r.instanceService.Save(toCreate)
	if err != nil {
		fmt.Errorf("### erro %s", err)
		c.Error(err)
		return
	}

	c.JSON(200, business.Response{
		Data: instance,
	})
}

func (r *instanceRouter) getInstances(c *gin.Context) {
	instances, err := r.instanceService.GetAll()
	if err != nil {
		fmt.Errorf("### erro %s", err)
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, business.Response{
		Data: instances,
	})
}

func (r *instanceRouter) getInstance(c *gin.Context) {
	id := c.Param("id")
	instanceFound, err := r.instanceService.Get(id)
	if err != nil {
		fmt.Errorf("### erro %s", err)
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, business.Response{
		Data: instanceFound,
	})
}

func (r *instanceRouter) deleteInstance(c *gin.Context) {
	id := c.Param("id")
	err := r.instanceService.Delete(id)
	if err != nil {
		fmt.Errorf("### erro %s", err)
		c.Error(err)
		return
	}

	c.JSON(http.StatusNoContent, nil)
}

func InstanceRouter(router gin.IRouter, instanceService services.InstanceService) routers.Router {
	r := &instanceRouter{
		instanceService: instanceService,
	}

	router.GET("", r.getInstances)
	router.GET("/:id", r.getInstance)
	router.DELETE("/:id", r.deleteInstance)
	router.POST("", r.postInstance)

	return r
}
