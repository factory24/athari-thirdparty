package onesignal

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/OneSignal/onesignal-go-api"
	"github.com/factory24/athari-thirdparty/pkg/data/dtos"
	"github.com/google/uuid"
)

type OneSignalClient interface {
	Connect()
	SendNotification(dto *dtos.PushNotification) error
}

type oneSignalClient struct {
	client    *onesignal.APIClient
	osAuthCtx context.Context
	appId     string
}

func (o *oneSignalClient) Connect() {
	o.appId = os.Getenv("ONESIGNAL.APP_ID")
	restApiKey := os.Getenv("ONESIGNAL.API_KEY")
	o.osAuthCtx = context.WithValue(
		context.Background(),
		onesignal.AppAuth,
		restApiKey,
	)
}

func (o *oneSignalClient) SendNotification(dto *dtos.PushNotification) error {

	notification := onesignal.NewNotification(o.appId)
	id := uuid.NewString()
	notification.SetId(id)
	if dto.To != "" {
		to := dto.To
		notification.SetExternalId(to)
		notification.SetIncludeExternalUserIds([]string{to})
	}
	stringMap := onesignal.StringMap{En: &dto.Content}
	notification.SetContents(stringMap)
	if dto.Title != "" {
		title := onesignal.StringMap{En: &dto.Title}
		notification.SetHeadings(title)
	}
	if dto.SubTitle != "" {
		subtitle := onesignal.StringMap{En: &dto.SubTitle}
		notification.SetSubtitle(subtitle)
	}
	if dto.Data != nil {
		notification.SetData(dto.Data)
	}

	notification.SetIsIos(false)

	resp, r, err := o.client.DefaultApi.CreateNotification(o.osAuthCtx).Notification(*notification).Execute()
	if r != nil && r.Body != nil {
		bodyBytes, readErr := io.ReadAll(r.Body)
		if readErr == nil {
			fmt.Printf("HTTP Response Body: %s\n", string(bodyBytes))
		}
		r.Body.Close()
	}
	if err != nil {
		fmt.Printf("Error when calling `CreateNotification`: %v\n", err)
		return err
	}
	fmt.Printf("Successfully created notification.\n")
	fmt.Printf("Notification ID: %v\n", resp.GetId())
	fmt.Printf("Notification Err: %v\n", resp.GetErrors())
	fmt.Printf("Notification Ext: %v\n", resp.GetExternalId())
	fmt.Printf("Notification reci: %v\n", resp.GetRecipients())

	return nil

}

func NewOneSignalClient() OneSignalClient {
	client := onesignal.NewAPIClient(onesignal.NewConfiguration())

	return &oneSignalClient{
		client: client,
	}
}
