package ecs_provisioner

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/dchest/uniuri"
	"go.uber.org/zap"

	"github.com/pushaas/pushaas/pushaas/models"
	"github.com/pushaas/pushaas/pushaas/provisioners"
)

type (
	ecsProvisioner struct {
		logger                *zap.Logger
		session               *session.Session
		provisionerConfig     *EcsProvisionerConfig
		pushRedisProvisioner  EcsPushRedisProvisioner
		pushStreamProvisioner EcsPushStreamProvisioner
		pushApiProvisioner    EcsPushApiProvisioner
	}
)

/*
	This whole provisioner is just a POC.
	It should be rewritten before using it in real code.

	The main bad points are:
		- the code was first written trying to parallelize steps (using channels to later synchronize), but things got difficult
		  and I just ended up running everything sequentially, but kept the channels in order to change as little as possible.
		- several points should consider adding load balancers to allow working with multiple instances, and using
		  names (instead of IPs) to address services
 */

func (p *ecsProvisioner) Provision(instance *models.Instance) *provisioners.PushServiceProvisionResult {
	p.logger.Info("starting provision for instance", zap.Any("instance", instance))

	var err error
	failureResult := &provisioners.PushServiceProvisionResult{
		Instance: instance,
		Status: provisioners.PushServiceProvisionStatusFailure,
		EnvVars: map[string]string{},
	}

	role, err := getIamRole(p.provisionerConfig.iam)
	if err != nil {
		p.logger.Error("failed while provisioning instance, failed to get iam role", zap.Any("instance", instance), zap.Error(err))
		return failureResult
	}

	/*
		push-redis
	*/
	chRedis := make(chan provisionPushRedisResult)
	go p.pushRedisProvisioner.Provision(instance, chRedis)
	resultPushRedis := <-chRedis
	if resultPushRedis.err != nil {
		p.logger.Error("push-redis: provision failure", zap.Any("instance", instance), zap.Error(resultPushRedis.err))
		// TODO deprovision
		return failureResult
	}
	p.logger.Info("push-redis: provision success", zap.Any("instance", instance))

	/*
		push-stream
	*/
	chStream := make(chan provisionPushStreamResult)
	go p.pushStreamProvisioner.Provision(instance, role, chStream)
	resultPushStream := <-chStream
	if resultPushStream.err != nil {
		p.logger.Error("push-stream: provision failure", zap.Any("instance", instance), zap.Error(resultPushStream.err))
		// TODO deprovision
		return failureResult
	}
	p.logger.Info("push-stream: provision success", zap.Any("instance", instance))

	/*
		push-api
	*/
	chApi := make(chan provisionPushApiResult)
	// TODO technical debt
	pushStreamPublicIp := *resultPushStream.eni.NetworkInterfaces[0].Association.PublicIp
	username := "app"
	password := uniuri.New()
	go p.pushApiProvisioner.Provision(instance, role, username, password, pushStreamPublicIp, chApi)
	resultPushApi := <-chApi
	if resultPushApi.err != nil {
		p.logger.Error("push-api: provision failure", zap.Any("instance", instance), zap.Error(resultPushApi.err))
		// TODO deprovision
		return failureResult
	}
	p.logger.Info("push-api: provision success", zap.Any("instance", instance))

	p.logger.Info(
		"finishing provision for instance",
		zap.Any("instance", instance),
		zap.Any("resultPushRedis", resultPushRedis),
		zap.Any("resultPushStream", resultPushStream),
		zap.Any("resultPushApi", resultPushApi),
	)

	// TODO technical debt
	pushApiPrivateIp := *resultPushApi.eni.NetworkInterfaces[0].PrivateIpAddress

	envVars := map[string]string{
		provisioners.EnvVarEndpoint: fmt.Sprintf("http://%s:%s", pushApiPrivateIp, pushApiPort),
		provisioners.EnvVarPassword: password,
		provisioners.EnvVarUsername: username,
	}

	return &provisioners.PushServiceProvisionResult{
		Instance: instance,
		EnvVars:  envVars,
		Status:   provisioners.PushServiceProvisionStatusSuccess,
	}
}

func (p *ecsProvisioner) Deprovision(instance *models.Instance) *provisioners.PushServiceDeprovisionResult {
	failureResult := &provisioners.PushServiceDeprovisionResult{
		Instance: instance,
		Status: provisioners.PushServiceDeprovisionStatusFailure,
	}

	/*
		push-api
	*/
	chApi := make(chan deprovisionPushApiResult)
	go p.pushApiProvisioner.Deprovision(instance, chApi)
	resultPushApi := <-chApi
	if resultPushApi.err != nil {
		p.logger.Error("push-api: deprovision failure", zap.Any("instance", instance), zap.Error(resultPushApi.err))
		return failureResult
	}
	p.logger.Info("push-api: deprovision success", zap.Any("instance", instance))

	/*
		push-stream
	*/
	chStream := make(chan deprovisionPushStreamResult)
	go p.pushStreamProvisioner.Deprovision(instance, chStream)
	resultPushStream := <-chStream
	if resultPushStream.err != nil {
		p.logger.Error("push-stream: deprovision failure", zap.Any("instance", instance), zap.Error(resultPushStream.err))
		return failureResult
	}
	p.logger.Info("push-stream: deprovision success", zap.Any("instance", instance))

	/*
		push-redis
	*/
	chRedis := make(chan deprovisionPushRedisResult)
	go p.pushRedisProvisioner.Deprovision(instance, chRedis)
	resultPushRedis := <-chRedis
	if resultPushRedis.err != nil {
		p.logger.Error("push-redis: deprovision failure", zap.Any("instance", instance), zap.Error(resultPushRedis.err))
		return failureResult
	}
	p.logger.Info("push-redis: deprovision success", zap.Any("instance", instance))

	p.logger.Info(
		"finishing deprovision for instance",
		zap.Any("instance", instance),
		zap.Any("resultPushRedis", resultPushRedis),
		zap.Any("resultPushStream", resultPushStream),
		zap.Any("resultPushApi", resultPushApi),
	)

	return &provisioners.PushServiceDeprovisionResult{
		Instance: instance,
		Status:   provisioners.PushServiceDeprovisionStatusSuccess,
	}
}

func NewEcsPushServiceProvisioner(
	logger *zap.Logger,
	provisionerConfig *EcsProvisionerConfig,
	pushRedisProvisioner EcsPushRedisProvisioner,
	pushStreamProvisioner EcsPushStreamProvisioner,
	pushApiProvisioner EcsPushApiProvisioner,
) (provisioners.PushServiceProvisioner, error) {
	return &ecsProvisioner{
		logger:                logger,
		provisionerConfig:     provisionerConfig,
		pushRedisProvisioner:  pushRedisProvisioner,
		pushStreamProvisioner: pushStreamProvisioner,
		pushApiProvisioner:    pushApiProvisioner,
	}, nil
}
