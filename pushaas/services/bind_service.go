package services

import (
	"go.uber.org/zap"

	"github.com/rafaeleyng/pushaas/pushaas/models"
)

type (
	AppBindResult   int
	AppUnbindResult int

	UnitBindResult   int
	UnitUnbindResult int

	BindService interface {
		BindApp(name string, bindAppForm *models.BindAppForm) (map[string]string, AppBindResult)
		UnbindApp(name string, bindAppForm *models.BindAppForm) AppUnbindResult
		BindUnit(name string, bindUnitForm *models.BindUnitForm) UnitBindResult
		UnbindUnit(name string, bindUnitForm *models.BindUnitForm) UnitUnbindResult
	}

	bindService struct {
		instanceService InstanceService
		logger          *zap.Logger
	}
)

const (
	AppBindSuccess AppBindResult = iota
	AppBindInstanceNotFound
	AppBindInstancePending
	AppBindInstanceFailed
	AppBindAlreadyBound
	AppBindFailure
)

const (
	AppUnbindSuccess AppUnbindResult = iota
	AppUnbindInstanceNotFound
	AppUnbindNotBound
	AppUnbindFailure
)

const (
	UnitBindSuccess UnitBindResult = iota
	UnitBindAlreadyBound
	UnitBindInstancePending
	UnitBindInstanceNotFound
	UnitBindFailure
)

const (
	UnitUnbindSuccess UnitUnbindResult = iota
	UnitUnbindAlreadyUnbound
	UnitUnbindInstanceNotFound
	UnitUnbindFailure
)

func findAppIndexInBindings(appName string, bindings []models.InstanceBinding) int {
	for i, binding := range bindings {
		if binding.AppName == appName {
			return i
		}
	}
	return -1
}

func findAppInBindings(appName string, bindings []models.InstanceBinding) *models.InstanceBinding {
	i := findAppIndexInBindings(appName, bindings)
	if i == -1 {
		return &models.InstanceBinding{}
	}

	return &bindings[i]
}

func (s *bindService) BindApp(name string, bindAppForm *models.BindAppForm) (map[string]string, AppBindResult) {
	panic("")
	//instance, result := s.instanceService.GetByName(name)
	//
	//if result == InstanceRetrievalNotFound {
	//	s.logger.Error("instance not found for binding", zap.String("name", name), zap.Any("bindAppForm", bindAppForm))
	//	return nil, AppBindInstanceNotFound
	//} else if instance.Status == models.InstanceStatusPending {
	//	return nil, AppBindInstancePending
	//} else if instance.Status == models.InstanceStatusFailed {
	//	return nil, AppBindInstanceFailed
	//}
	//
	//instanceBinding := findAppInBindings(bindAppForm.AppName, instance.Bindings)
	//if instanceBinding.AppName == bindAppForm.AppName {
	//	return nil, AppBindAlreadyBound
	//}
	//
	//instanceBinding = &models.InstanceBinding{
	//	AppName: bindAppForm.AppName,
	//	AppHost: bindAppForm.AppHost,
	//}
	//instance.Bindings = append(instance.Bindings, *instanceBinding)
	//err := s.instanceService.GetCollection().Save(instance)
	//if err != nil {
	//	s.logger.Error("failed to bind to instance", zap.String("name", name), zap.Any("bindAppForm", bindAppForm), zap.Any("instance", instance))
	//	return nil, AppBindFailure
	//}
	//
	//envVars := map[string]string{
	//	"PUSHAAS_ENDPOINT": "TODO-endpoint",
	//	"PUSHAAS_USERNAME": "TODO-username",
	//	"PUSHAAS_PASSWORD": "TODO-password",
	//}
	//
	//return envVars, AppBindSuccess
}

func (s *bindService) UnbindApp(name string, bindAppForm *models.BindAppForm) AppUnbindResult {
	panic("")
	//instance, result := s.instanceService.GetByName(name)
	//
	//if result == InstanceRetrievalNotFound {
	//	s.logger.Error("instance not found for unbinding", zap.String("name", name), zap.Any("bindAppForm", bindAppForm))
	//	return AppUnbindInstanceNotFound
	//}
	//
	//i := findAppIndexInBindings(bindAppForm.AppName, instance.Bindings)
	//if i == -1 {
	//	s.logger.Error("instance is not bound to app", zap.String("name", name), zap.Any("bindAppForm", bindAppForm))
	//	return AppUnbindNotBound
	//}
	//
	//instance.Bindings = append(instance.Bindings[:i], instance.Bindings[i+1:]...)
	//err := s.instanceService.GetCollection().Save(instance)
	//if err != nil {
	//	s.logger.Error("failed to unbind to instance", zap.String("name", name), zap.Any("bindAppForm", bindAppForm), zap.Any("instance", instance))
	//	return AppUnbindFailure
	//}
	//
	//return AppUnbindSuccess
}

func (s *bindService) BindUnit(name string, bindUnitForm *models.BindUnitForm) UnitBindResult {
	// TODO implement
	panic("implement me")
}

func (s *bindService) UnbindUnit(name string, bindUnitForm *models.BindUnitForm) UnitUnbindResult {
	// TODO implement
	panic("implement me")
}

func NewBindService(logger *zap.Logger, instanceService InstanceService) BindService {
	return &bindService{
		instanceService: instanceService,
		logger:          logger,
	}
}
