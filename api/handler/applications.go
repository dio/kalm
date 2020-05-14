package handler

import (
	"encoding/json"
	"github.com/kapp-staging/kapp/api/resources"
	"github.com/kapp-staging/kapp/controller/api/v1alpha1"
	"github.com/labstack/echo/v4"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"net/http"
)

func (h *ApiHandler) handleGetApplications(c echo.Context) error {
	applicationList, err := getKappApplicationList(c)

	if err != nil {
		return err
	}

	res, err := h.applicationListResponse(c, applicationList)

	if err != nil {
		return err
	}

	return c.JSON(200, res)
}

func (h *ApiHandler) handleGetApplicationDetails(c echo.Context) error {
	application, err := getKappApplication(c)

	if err != nil {
		return err
	}

	res, err := h.applicationResponse(c, application)

	if err != nil {
		return err
	}

	return c.JSON(200, res)
}

func (h *ApiHandler) handleValidateApplications(c echo.Context) error {
	crdApplication, _, err := getApplicationFromContext(c)
	if err != nil {
		return err
	}

	if err := v1alpha1.TryValidateApplicationFromAPI(crdApplication.Spec, crdApplication.Name); err != nil {
		return err
	}

	return nil
}

func (h *ApiHandler) handleCreateApplication(c echo.Context) error {
	application, err := createKappApplication(c)

	if err != nil {
		return err
	}

	res, err := h.applicationResponse(c, application)

	if err != nil {
		return err
	}

	return c.JSON(http.StatusCreated, res)
}

func (h *ApiHandler) handleUpdateApplication(c echo.Context) error {
	application, err := updateKappApplication(c)

	if err != nil {
		return err
	}

	res, err := h.applicationResponse(c, application)

	if err != nil {
		return err
	}

	return c.JSON(200, res)
}

func (h *ApiHandler) handleDeleteApplication(c echo.Context) error {
	err := deleteKappApplication(c)
	if err != nil {
		return err
	}
	return c.NoContent(http.StatusNoContent)
}

// Helper functions

func deleteKappApplication(c echo.Context) error {
	k8sClient := getK8sClient(c)
	_, err := k8sClient.RESTClient().Delete().Body(c.Request().Body).AbsPath(kappApplicationUrl(c)).DoRaw()
	return err
}

func createKappApplication(c echo.Context) (*v1alpha1.Application, error) {
	k8sClient := getK8sClient(c)

	crdApplication, plugins, err := getApplicationFromContext(c)
	if err != nil {
		return nil, err
	}

	if err := v1alpha1.TryValidateApplicationFromAPI(crdApplication.Spec, crdApplication.Name); err != nil {
		return nil, err
	}

	bts, _ := json.Marshal(crdApplication)
	var application v1alpha1.Application
	err = k8sClient.RESTClient().Post().Body(bts).AbsPath(kappApplicationUrl(c)).Do().Into(&application)
	if err != nil {
		return nil, err
	}

	kappClient, _ := getKappV1Alpha1Client(c)
	err = resources.UpdateApplicationPluginBindingsForObject(kappClient, application.Name, "", plugins)

	if err != nil {
		return nil, err
	}

	return &application, nil
}

func updateKappApplication(c echo.Context) (*v1alpha1.Application, error) {
	k8sClient := getK8sClient(c)

	crdApplication, plugins, err := getApplicationFromContext(c)

	if err != nil {
		return nil, err
	}

	fetched, err := getKappApplication(c)

	if err != nil {
		return nil, err
	}
	crdApplication.ResourceVersion = fetched.ResourceVersion

	bts, _ := json.Marshal(crdApplication)
	var application v1alpha1.Application
	err = k8sClient.RESTClient().Put().Body(bts).AbsPath(kappApplicationUrl(c)).Do().Into(&application)

	if err != nil {
		return nil, err
	}

	kappClient, _ := getKappV1Alpha1Client(c)
	err = resources.UpdateApplicationPluginBindingsForObject(kappClient, application.Name, "", plugins)

	if err != nil {
		return nil, err
	}

	return &application, nil
}

func getKappApplication(c echo.Context) (*v1alpha1.Application, error) {
	k8sClient := getK8sClient(c)
	var fetched v1alpha1.Application
	err := k8sClient.RESTClient().Get().AbsPath(kappApplicationUrl(c)).Do().Into(&fetched)
	if err != nil {
		return nil, err
	}
	return &fetched, nil
}

func getKappApplicationList(c echo.Context) (*v1alpha1.ApplicationList, error) {
	k8sClient := getK8sClient(c)
	var fetched v1alpha1.ApplicationList
	err := k8sClient.RESTClient().Get().AbsPath(kappApplicationUrl(c)).Do().Into(&fetched)
	if err != nil {
		return nil, err
	}
	return &fetched, nil
}

func kappApplicationUrl(c echo.Context) string {
	name := c.Param("name")

	if name == "" {
		return "/apis/core.kapp.dev/v1alpha1/applications"
	}

	return "/apis/core.kapp.dev/v1alpha1/applications/" + name
}

func getApplicationFromContext(c echo.Context) (*v1alpha1.Application, []runtime.RawExtension, error) {
	var application resources.Application

	if err := c.Bind(&application); err != nil {
		return nil, nil, err
	}

	crdApplication := &v1alpha1.Application{
		TypeMeta: metaV1.TypeMeta{
			Kind:       "Application",
			APIVersion: "core.kapp.dev/v1alpha1",
		},
		ObjectMeta: metaV1.ObjectMeta{
			Name: application.Name,
		},
		Spec: v1alpha1.ApplicationSpec{
			IsActive:  application.IsActive,
			SharedEnv: application.SharedEnvs,
			//PluginsNew: application.Plugins,
		},
	}

	return crdApplication, application.Plugins, nil
}

func (h *ApiHandler) applicationResponse(c echo.Context, application *v1alpha1.Application) (*resources.ApplicationDetails, error) {
	k8sClient := getK8sClient(c)
	k8sClientConfig := getK8sClientConfig(c)

	builder := resources.Builder{
		K8sClient: k8sClient,
		Logger:    h.logger,
		Config:    k8sClientConfig,
	}

	return builder.BuildApplicationDetails(application)
}

func (h *ApiHandler) applicationListResponse(c echo.Context, applicationList *v1alpha1.ApplicationList) ([]resources.ApplicationDetails, error) {
	k8sClient := getK8sClient(c)
	k8sClientConfig := getK8sClientConfig(c)

	builder := resources.Builder{
		K8sClient: k8sClient,
		Logger:    h.logger,
		Config:    k8sClientConfig,
	}

	return builder.BuildApplicationListResponse(applicationList)
}
