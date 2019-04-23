package v1

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/rafaeleyng/pushaas/pushaas/models"

	"github.com/rafaeleyng/pushaas/pushaas/business"
)

//func handleGetServices(c *gin.Context) {
//	c.JSON(http.StatusOK, business.Response{
//		Data: gin.H{
//			"TODO": "TODO",
//		},
//	})
//}

//func handlePostServices(c *gin.Context) {
//	// sess, err := session.NewSession()
//	// if err != nil {
//	// 	fmt.Errorf("err", err)
//	// 	return
//	// }
//
//	// svc := ecs.New(sess)
//
//	// CreateService
//	// RegisterTaskDefinition
//	// RunTask | StartTask
//
//}

func handlePostInstance(c *gin.Context) {
	// sess, err := session.NewSession()
	// if err != nil {
	// 	fmt.Errorf("err", err)
	// 	return
	// }

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

	toCreate := &models.Instance{
		Description: "an instance of a push service",
	}

	instance, err := models.InstanceSave(toCreate)
	if err != nil {
		fmt.Errorf("### erro %s", err)
		c.Error(err)
		return
	}

	c.JSON(200, business.Response{
		Data: instance,
	})
}

func handleGetInstances(c *gin.Context) {
	instances, err := models.InstanceGetAll()
	if err != nil {
		fmt.Errorf("### erro %s", err)
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, business.Response{
		Data: instances,
	})
}

func handleGetInstance(c *gin.Context) {
	id := c.Param("id")
	instance, err := models.InstanceGet(id)
	if err != nil {
		fmt.Errorf("### erro %s", err)
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, business.Response{
		Data: instance,
	})
}

func handleDeleteInstance(c *gin.Context) {
	id := c.Param("id")
	err := models.InstanceDelete(id)
	if err != nil {
		fmt.Errorf("### erro %s", err)
		c.Error(err)
		return
	}

	c.JSON(http.StatusNoContent, nil)
}
