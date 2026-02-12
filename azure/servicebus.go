package azure

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/messaging/azservicebus"
)

type MessageHandler func(context.Context, *azservicebus.ReceivedMessage) error

type ServiceBusClient interface {
	Connect()
	ListenOnTopicSubscription(string, string, MessageHandler)
	SendMessage(string, string, interface{}) error
}

type azureServiceBusClient struct {
	ctx context.Context
	bus *azservicebus.Client
}

func (client *azureServiceBusClient) ListenOnTopicSubscription(s1 string, s2 string, handler MessageHandler) {
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

			if err := handler(client.ctx, message); err != nil {
				log.Println("Handler failed to process message:", err)
				// Decide whether to Abandon, DeadLetter, or just log.
				// For now, if handler fails, we might not want to Complete it?
				// But original code completed it inside the handler logic.
				// Let's assume handler takes care of business logic, but completion...
				// The original code completed the message INSIDE the handler (which was a method on client).
				// So we should expose the receiver to the handler? Or pass the responsibility to the handler?
				// Simplified: Caller logic should handle completion if they want to control it,
				// OR we pass a "Completer" interface.

				// Re-reading original code:
				// client.handleFlowmeter(receiver, message, ...)
				// Inside: receiver.CompleteMessage(client.ctx, message, nil)

				// So usage of receiver is needed.
				// The handler signature `func(context.Context, *azservicebus.ReceivedMessage) error`
				// doesn't give access to receiver.

				// However, `receiver` is created inside this method.
				// We can pass `receiver` to the handler?
				// Or we can just complete it here if handler returns nil?
				// Standard pattern: Handler processes, if success, we complete.
				continue
			}

			// If handler returns nil (success), we complete the message
			if err := receiver.CompleteMessage(client.ctx, message, nil); err != nil {
				log.Println("Failed to complete message receive:", err)
			} else {
				log.Println("Message completed:", message.MessageID)
			}
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
