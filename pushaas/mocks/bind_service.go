// Code generated by moq; DO NOT EDIT.
// github.com/matryer/moq

package mocks

import (
	"github.com/rafaeleyng/pushaas/pushaas/models"
	"github.com/rafaeleyng/pushaas/pushaas/services"
	"sync"
)

var (
	lockBindServiceMockBindApp    sync.RWMutex
	lockBindServiceMockBindUnit   sync.RWMutex
	lockBindServiceMockUnbindApp  sync.RWMutex
	lockBindServiceMockUnbindUnit sync.RWMutex
)

// Ensure, that BindServiceMock does implement BindService.
// If this is not the case, regenerate this file with moq.
var _ services.BindService = &BindServiceMock{}

// BindServiceMock is a mock implementation of BindService.
//
//     func TestSomethingThatUsesBindService(t *testing.T) {
//
//         // make and configure a mocked BindService
//         mockedBindService := &BindServiceMock{
//             BindAppFunc: func(name string, bindAppForm *models.BindAppForm) (map[string]string, services.AppBindResult) {
// 	               panic("mock out the BindApp method")
//             },
//             BindUnitFunc: func(name string, bindUnitForm *models.BindUnitForm) services.UnitBindResult {
// 	               panic("mock out the BindUnit method")
//             },
//             UnbindAppFunc: func(name string, bindAppForm *models.BindAppForm) services.AppUnbindResult {
// 	               panic("mock out the UnbindApp method")
//             },
//             UnbindUnitFunc: func(name string, bindUnitForm *models.BindUnitForm) services.UnitUnbindResult {
// 	               panic("mock out the UnbindUnit method")
//             },
//         }
//
//         // use mockedBindService in code that requires BindService
//         // and then make assertions.
//
//     }
type BindServiceMock struct {
	// BindAppFunc mocks the BindApp method.
	BindAppFunc func(name string, bindAppForm *models.BindAppForm) (map[string]string, services.AppBindResult)

	// BindUnitFunc mocks the BindUnit method.
	BindUnitFunc func(name string, bindUnitForm *models.BindUnitForm) services.UnitBindResult

	// UnbindAppFunc mocks the UnbindApp method.
	UnbindAppFunc func(name string, bindAppForm *models.BindAppForm) services.AppUnbindResult

	// UnbindUnitFunc mocks the UnbindUnit method.
	UnbindUnitFunc func(name string, bindUnitForm *models.BindUnitForm) services.UnitUnbindResult

	// calls tracks calls to the methods.
	calls struct {
		// BindApp holds details about calls to the BindApp method.
		BindApp []struct {
			// Name is the name argument value.
			Name string
			// BindAppForm is the bindAppForm argument value.
			BindAppForm *models.BindAppForm
		}
		// BindUnit holds details about calls to the BindUnit method.
		BindUnit []struct {
			// Name is the name argument value.
			Name string
			// BindUnitForm is the bindUnitForm argument value.
			BindUnitForm *models.BindUnitForm
		}
		// UnbindApp holds details about calls to the UnbindApp method.
		UnbindApp []struct {
			// Name is the name argument value.
			Name string
			// BindAppForm is the bindAppForm argument value.
			BindAppForm *models.BindAppForm
		}
		// UnbindUnit holds details about calls to the UnbindUnit method.
		UnbindUnit []struct {
			// Name is the name argument value.
			Name string
			// BindUnitForm is the bindUnitForm argument value.
			BindUnitForm *models.BindUnitForm
		}
	}
}

// BindApp calls BindAppFunc.
func (mock *BindServiceMock) BindApp(name string, bindAppForm *models.BindAppForm) (map[string]string, services.AppBindResult) {
	if mock.BindAppFunc == nil {
		panic("BindServiceMock.BindAppFunc: method is nil but BindService.BindApp was just called")
	}
	callInfo := struct {
		Name        string
		BindAppForm *models.BindAppForm
	}{
		Name:        name,
		BindAppForm: bindAppForm,
	}
	lockBindServiceMockBindApp.Lock()
	mock.calls.BindApp = append(mock.calls.BindApp, callInfo)
	lockBindServiceMockBindApp.Unlock()
	return mock.BindAppFunc(name, bindAppForm)
}

// BindAppCalls gets all the calls that were made to BindApp.
// Check the length with:
//     len(mockedBindService.BindAppCalls())
func (mock *BindServiceMock) BindAppCalls() []struct {
	Name        string
	BindAppForm *models.BindAppForm
} {
	var calls []struct {
		Name        string
		BindAppForm *models.BindAppForm
	}
	lockBindServiceMockBindApp.RLock()
	calls = mock.calls.BindApp
	lockBindServiceMockBindApp.RUnlock()
	return calls
}

// BindUnit calls BindUnitFunc.
func (mock *BindServiceMock) BindUnit(name string, bindUnitForm *models.BindUnitForm) services.UnitBindResult {
	if mock.BindUnitFunc == nil {
		panic("BindServiceMock.BindUnitFunc: method is nil but BindService.BindUnit was just called")
	}
	callInfo := struct {
		Name         string
		BindUnitForm *models.BindUnitForm
	}{
		Name:         name,
		BindUnitForm: bindUnitForm,
	}
	lockBindServiceMockBindUnit.Lock()
	mock.calls.BindUnit = append(mock.calls.BindUnit, callInfo)
	lockBindServiceMockBindUnit.Unlock()
	return mock.BindUnitFunc(name, bindUnitForm)
}

// BindUnitCalls gets all the calls that were made to BindUnit.
// Check the length with:
//     len(mockedBindService.BindUnitCalls())
func (mock *BindServiceMock) BindUnitCalls() []struct {
	Name         string
	BindUnitForm *models.BindUnitForm
} {
	var calls []struct {
		Name         string
		BindUnitForm *models.BindUnitForm
	}
	lockBindServiceMockBindUnit.RLock()
	calls = mock.calls.BindUnit
	lockBindServiceMockBindUnit.RUnlock()
	return calls
}

// UnbindApp calls UnbindAppFunc.
func (mock *BindServiceMock) UnbindApp(name string, bindAppForm *models.BindAppForm) services.AppUnbindResult {
	if mock.UnbindAppFunc == nil {
		panic("BindServiceMock.UnbindAppFunc: method is nil but BindService.UnbindApp was just called")
	}
	callInfo := struct {
		Name        string
		BindAppForm *models.BindAppForm
	}{
		Name:        name,
		BindAppForm: bindAppForm,
	}
	lockBindServiceMockUnbindApp.Lock()
	mock.calls.UnbindApp = append(mock.calls.UnbindApp, callInfo)
	lockBindServiceMockUnbindApp.Unlock()
	return mock.UnbindAppFunc(name, bindAppForm)
}

// UnbindAppCalls gets all the calls that were made to UnbindApp.
// Check the length with:
//     len(mockedBindService.UnbindAppCalls())
func (mock *BindServiceMock) UnbindAppCalls() []struct {
	Name        string
	BindAppForm *models.BindAppForm
} {
	var calls []struct {
		Name        string
		BindAppForm *models.BindAppForm
	}
	lockBindServiceMockUnbindApp.RLock()
	calls = mock.calls.UnbindApp
	lockBindServiceMockUnbindApp.RUnlock()
	return calls
}

// UnbindUnit calls UnbindUnitFunc.
func (mock *BindServiceMock) UnbindUnit(name string, bindUnitForm *models.BindUnitForm) services.UnitUnbindResult {
	if mock.UnbindUnitFunc == nil {
		panic("BindServiceMock.UnbindUnitFunc: method is nil but BindService.UnbindUnit was just called")
	}
	callInfo := struct {
		Name         string
		BindUnitForm *models.BindUnitForm
	}{
		Name:         name,
		BindUnitForm: bindUnitForm,
	}
	lockBindServiceMockUnbindUnit.Lock()
	mock.calls.UnbindUnit = append(mock.calls.UnbindUnit, callInfo)
	lockBindServiceMockUnbindUnit.Unlock()
	return mock.UnbindUnitFunc(name, bindUnitForm)
}

// UnbindUnitCalls gets all the calls that were made to UnbindUnit.
// Check the length with:
//     len(mockedBindService.UnbindUnitCalls())
func (mock *BindServiceMock) UnbindUnitCalls() []struct {
	Name         string
	BindUnitForm *models.BindUnitForm
} {
	var calls []struct {
		Name         string
		BindUnitForm *models.BindUnitForm
	}
	lockBindServiceMockUnbindUnit.RLock()
	calls = mock.calls.UnbindUnit
	lockBindServiceMockUnbindUnit.RUnlock()
	return calls
}
