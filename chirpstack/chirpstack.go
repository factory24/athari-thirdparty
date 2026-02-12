package chirpstack

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/chirpstack/chirpstack/api/go/v4/api"
	"github.com/factory24/athari-thirdparty/pkg/data/dtos"
	"google.golang.org/grpc"
)

const (
	DownLinkPort = 10
)

type APIToken string

func (a APIToken) GetRequestMetadata(ctx context.Context, url ...string) (map[string]string, error) {
	return map[string]string{
		"authorization": fmt.Sprintf("Bearer %s", a),
	}, nil
}

func (a APIToken) RequireTransportSecurity() bool {
	return false
}

type NetworkServerClient interface {
	Connect()
	CreateDevice(dto dtos.DeviceDTO) error
	CreateGateway(dto dtos.GatewayDTO) error
	Enqueue(context.Context, *api.EnqueueDeviceQueueItemRequest) (*api.EnqueueDeviceQueueItemResponse, error)
	GetDevice(eui string) (*api.Device, error)
	GetGateway(context.Context, string) (*api.Gateway, *time.Time, error)
	GetQueue(ctx context.Context, request *api.GetDeviceQueueItemsRequest) (*api.GetDeviceQueueItemsResponse, error)
	SetKey(eui, key string) error
	UpdateDevice(dto dtos.DeviceDTO) error
	UpdateGateway(dto dtos.GatewayDTO) error
}

type chirpstackClient struct {
	deviceClient               api.DeviceServiceClient
	applicationServiceClient   api.ApplicationServiceClient
	gatewayServiceClient       api.GatewayServiceClient
	tenantServiceClient        api.TenantServiceClient
	deviceProfileServiceClient api.DeviceProfileServiceClient
	conn                       *grpc.ClientConn
}

func (client *chirpstackClient) GetQueue(ctx context.Context, request *api.GetDeviceQueueItemsRequest) (*api.GetDeviceQueueItemsResponse, error) {
	response, err := client.deviceClient.GetQueue(ctx, request)
	if err != nil {
		return nil, err
	}
	return response, nil
}

func (client *chirpstackClient) Connect() {
	log.Println("connecting to chirpstack server ...")
	dialOpts := []grpc.DialOption{
		grpc.WithBlock(),
		grpc.WithPerRPCCredentials(APIToken(os.Getenv("CS.API_KEY"))),
		grpc.WithInsecure(),
	}

	baseUrl := fmt.Sprintf("%s:%s", os.Getenv("CS.BASE_URL"), os.Getenv("CS.PORT"))
	conn, err := grpc.Dial(baseUrl, dialOpts...)
	if err != nil {
		log.Println("failed to connect to chirpstack server ::::: |", err)
		os.Exit(3)
	}
	log.Println("connected to chirpstack server")

	client.deviceClient = api.NewDeviceServiceClient(conn)
	client.applicationServiceClient = api.NewApplicationServiceClient(conn)
	client.gatewayServiceClient = api.NewGatewayServiceClient(conn)
	client.tenantServiceClient = api.NewTenantServiceClient(conn)
	client.deviceProfileServiceClient = api.NewDeviceProfileServiceClient(conn)

	log.Println("connection to chirpstack server was successful")
}
func (client *chirpstackClient) GetDevice(eui string) (*api.Device, error) {
	request := &api.GetDeviceRequest{
		DevEui: eui,
	}
	res, err := client.deviceClient.Get(context.Background(), request)
	if err != nil {
		return nil, err
	}
	return res.GetDevice(), nil
}

func (client *chirpstackClient) CreateDevice(dto dtos.DeviceDTO) error {
	request := &api.CreateDeviceRequest{
		Device: &api.Device{
			DevEui:          dto.Eui,
			Name:            dto.Name,
			Description:     dto.Description,
			ApplicationId:   os.Getenv("CS.APPLICATION_ID"),
			DeviceProfileId: os.Getenv("CS.DEVICE_PROFILE_ID"),
			SkipFcntCheck:   true,
			IsDisabled:      false,
			Variables:       nil,
			Tags:            nil,
			JoinEui:         os.Getenv("CS.LORA_JOIN_EUI"),
		},
	}

	_, err := client.deviceClient.Create(context.Background(), request)
	if err != nil {
		log.Println("error creating device", err)
		return err
	}

	return err
}

func (client *chirpstackClient) CreateGateway(dto dtos.GatewayDTO) error {
	gateway := &api.CreateGatewayRequest{
		Gateway: &api.Gateway{
			GatewayId:     dto.SerialNumber,
			Name:          dto.Name,
			Description:   dto.Description,
			TenantId:      os.Getenv("CS.TENANT_ID"),
			Location:      dto.Location,
			Tags:          nil,
			Metadata:      nil,
			StatsInterval: dto.StatsInterval,
		},
	}
	_, err := client.gatewayServiceClient.Create(context.Background(), gateway)
	if err != nil {
		return err
	}
	return nil
}

func (client *chirpstackClient) GetGateway(ctx context.Context, serialNumber string) (*api.Gateway, *time.Time, error) {
	get, err := client.gatewayServiceClient.Get(ctx, &api.GetGatewayRequest{
		GatewayId: serialNumber,
	})
	if err != nil {
		return nil, nil, err
	}

	log.Println("gateway", get)
	asTime := get.GetLastSeenAt().AsTime()
	return get.GetGateway(), &asTime, nil
}

func (client *chirpstackClient) UpdateDevice(dto dtos.DeviceDTO) error {
	request := &api.UpdateDeviceRequest{
		Device: &api.Device{
			DevEui:          dto.Eui,
			Name:            dto.Name,
			Description:     dto.Description,
			ApplicationId:   os.Getenv("CS.APPLICATION_ID"),
			DeviceProfileId: os.Getenv("CS.DEVICE_PROFILE_ID"),
			SkipFcntCheck:   true,
			IsDisabled:      false,
			Variables:       nil,
			Tags:            nil,
			JoinEui:         os.Getenv("CS.LORA_JOIN_EUI"),
		},
	}
	_, err := client.deviceClient.Update(context.Background(), request)
	if err != nil {
		log.Println("error updating device", err)
		return err
	}
	return nil
}

func (client *chirpstackClient) UpdateGateway(dto dtos.GatewayDTO) error {
	gateway := &api.UpdateGatewayRequest{
		Gateway: &api.Gateway{
			GatewayId:     dto.SerialNumber,
			Name:          dto.Name,
			Description:   dto.Description,
			TenantId:      os.Getenv("CS.TENANT_ID"),
			Location:      dto.Location,
			Tags:          nil,
			Metadata:      nil,
			StatsInterval: dto.StatsInterval,
		},
	}
	_, err := client.gatewayServiceClient.Update(context.Background(), gateway)
	if err != nil {
		return err
	}

	return nil
}

func (client *chirpstackClient) SetKey(eui, key string) error {

	_, err := client.deviceClient.GetKeys(context.Background(), &api.GetDeviceKeysRequest{
		DevEui: eui,
	})

	if err == nil {
		_, err = client.deviceClient.UpdateKeys(context.Background(), &api.UpdateDeviceKeysRequest{
			DeviceKeys: &api.DeviceKeys{
				DevEui: eui,
				AppKey: key,
				NwkKey: key,
			},
		})
		if err != nil {
			log.Println("error creating device keys", err)
			return err
		}
		return nil
	}
	_, err = client.deviceClient.CreateKeys(context.Background(), &api.CreateDeviceKeysRequest{
		DeviceKeys: &api.DeviceKeys{
			DevEui: eui,
			AppKey: key,
			NwkKey: key,
		},
	})
	if err != nil {
		log.Println("error creating device keys", err)
	}

	return err
}

func (client *chirpstackClient) Enqueue(
	ctx context.Context,
	request *api.EnqueueDeviceQueueItemRequest,
) (*api.EnqueueDeviceQueueItemResponse, error) {
	response, err := client.deviceClient.Enqueue(ctx, request)
	if err != nil {
		return nil, err
	}

	log.Println("Response :::::: |", response.String())

	return response, nil
}

func NewChirpstackClient() NetworkServerClient {
	return &chirpstackClient{}
}
