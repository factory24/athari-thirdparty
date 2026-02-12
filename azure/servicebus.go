package azure

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/messaging/azservicebus"
	"github.com/factory24/athari-flowbox-device-service/pkg/data/dtos"
	"github.com/factory24/athari-flowbox-device-service/pkg/data/models"
	"github.com/factory24/athari-flowbox-device-service/pkg/data/repositories"
)

type ServiceBusClient interface {
	Connect()
	ListenOnTopicSubscription(string, string, repositories.DeviceAttributeRepository)
	SendMessage(string, string, interface{}) error
}

type azureServiceBusClient struct {
	ctx context.Context
	bus *azservicebus.Client
}

func (client *azureServiceBusClient) ListenOnTopicSubscription(s1 string, s2 string, deviceAttributeRepository repositories.DeviceAttributeRepository) {
	log.Println("Azure service bus client for device attribute started...")
	receiver, err := client.bus.NewReceiverForSubscription(s1, s2, &azservicebus.ReceiverOptions{})
	if err != nil {
		log.Fatalln("Failed to create receiver for topic & subscription:", err)
	}

	log.Println("Azure service bus client receiver started...")
	for {
		messages, err := receiver.ReceiveMessages(client.ctx, 1, &azservicebus.ReceiveMessagesOptions{})
		if err != nil {
			log.Fatalln("Failed to receive messages:", err)
		}

		for _, message := range messages {
			log.Printf("Message received -> %v\n", string(message.Body))
			log.Println()

			command := new(dtos.TopicCommand)
			if err := json.Unmarshal(message.Body, command); err != nil {
				log.Println("Failed to unmarshal bytes into struct:", err)
			}

			client.handleFlowmeter(receiver, message, command.Payload.Logs.Flowmeter, deviceAttributeRepository)
			client.handleFlowbox(receiver, message, command.Payload.Logs.Flowbox, deviceAttributeRepository)
		}

		time.Sleep(2 * time.Second)
	}
}

func (client *azureServiceBusClient) Connect() {
	cl, err := azservicebus.NewClientFromConnectionString(os.Getenv("AZ.SB.CONNECTION_STRING"), nil)
	if err != nil {
		log.Println("failed to connect to azure service bus:", err)
		os.Exit(1)
	}

	log.Println("Connection to azure service bus was successful")
	client.bus = cl
}

func (client *azureServiceBusClient) handleFlowmeter(
	receiver *azservicebus.Receiver,
	message *azservicebus.ReceivedMessage,
	flowmeter *dtos.Flowmeter,
	deviceAttributeRepository repositories.DeviceAttributeRepository,
) {
	creditBalance := strconv.FormatFloat(flowmeter.Balance, 'f', 2, 64)
	balanceDto := &models.DeviceAttribute{
		SerialNumber: flowmeter.SerialNumber,
		Name:         "creditBalance",
		Value:        creditBalance,
	}

	consumptionBalance := strconv.FormatFloat(flowmeter.Consumption, 'f', 2, 64)
	consumptionDto := &models.DeviceAttribute{
		SerialNumber: flowmeter.SerialNumber,
		Name:         "creditConsumption",
		Value:        consumptionBalance,
	}

	mainBattery := strconv.Itoa(flowmeter.MainBatteryLevel)
	batteryDto := &models.DeviceAttribute{
		SerialNumber: flowmeter.SerialNumber,
		Name:         "mainBatteryLevel",
		Value:        mainBattery,
	}

	valveBattery := strconv.Itoa(flowmeter.ValveBatteryLevel)
	valveBatteryDto := &models.DeviceAttribute{
		SerialNumber: flowmeter.SerialNumber,
		Name:         "valveBatteryLevel",
		Value:        valveBattery,
	}

	deviceAttributes := []*models.DeviceAttribute{balanceDto, consumptionDto, batteryDto, valveBatteryDto}
	if err := deviceAttributeRepository.SaveOrUpdateAll(deviceAttributes); err != nil {
		log.Println("Failed to save device attributes ::::::: |", err)
		log.Println("Device attributes ::::::: |")
	}

	log.Println("Device attributes created:", deviceAttributes)
	if err := receiver.CompleteMessage(client.ctx, message, nil); err != nil {
		log.Fatalln("Failed to complete message receive:", err)
	}

	log.Println("Message completed:", message.MessageID)
}

func (client *azureServiceBusClient) handleFlowbox(
	receiver *azservicebus.Receiver,
	message *azservicebus.ReceivedMessage,
	flowbox *dtos.Flowbox,
	deviceAttributeRepository repositories.DeviceAttributeRepository,
) {
	cardCount := strconv.Itoa(flowbox.CardCount)
	cardCountDto := &models.DeviceAttribute{
		SerialNumber: flowbox.GatewayID,
		Name:         "cardCount",
		Value:        cardCount,
	}

	transactionCount := strconv.Itoa(flowbox.TransactionCount)
	transactionCountDto := &models.DeviceAttribute{
		SerialNumber: flowbox.GatewayID,
		Name:         "transactionCount",
		Value:        transactionCount,
	}

	topUpCount := strconv.Itoa(flowbox.TopUpCount)
	topUpCountDto := &models.DeviceAttribute{
		SerialNumber: flowbox.GatewayID,
		Name:         "topUpCount",
		Value:        topUpCount,
	}

	deviceAttributes := []*models.DeviceAttribute{cardCountDto, transactionCountDto, topUpCountDto}
	if err := deviceAttributeRepository.SaveOrUpdateAll(deviceAttributes); err != nil {
		log.Println("Failed to save device attributes ::::::: |", err)
	}

	log.Println("Device attributes created:", deviceAttributes)
	if err := receiver.CompleteMessage(client.ctx, message, nil); err != nil {
		log.Fatalln("Failed to complete message receive:", err)
	}

	log.Println("Message completed:", message.MessageID)
}

func (client *azureServiceBusClient) SendMessage(topic, event string, data interface{}) error {
	sender, err := client.bus.NewSender(os.Getenv("AZ.SB.QUEUE_NAME"), &azservicebus.NewSenderOptions{})
	if err != nil {
		return nil
	}

	jb, err := json.Marshal(data)
	if err != nil {
		return err
	}

	message := &azservicebus.Message{
		Body:    jb,
		Subject: &event,
		ApplicationProperties: map[string]any{
			"event": event,
		},
	}

	log.Printf("<- Message sending : %s->\n", data)
	if err = sender.SendMessage(client.ctx, message, &azservicebus.SendMessageOptions{}); err != nil {
		log.Println("<- failed send :", err)
		return err
	}

	log.Printf("<- Message sent")
	return nil
}

func NewAzureServiceBusClient(context context.Context) ServiceBusClient {
	return &azureServiceBusClient{
		ctx: context,
	}
}
