package dispatcher

import (
	"context"
	"encoding/json"
	"fmt"
	"slices"
	"sync"
	"time"

	"github.com/IsaqueB/notification-server/internal/broker"
	"github.com/IsaqueB/notification-server/internal/models"
	"github.com/IsaqueB/notification-server/pkg/logger"
)

var (
	ErrClientRequired            = fmt.Errorf("client is required")
	ErrClientIdRequired          = fmt.Errorf("client id is required")
	ErrClientSendChannelRequired = fmt.Errorf("client send is required")
	ErrTopicNonExistent          = fmt.Errorf("topic does not exist")
	ErrTopicRequired             = fmt.Errorf("topic is required")
)

type Dispatcher struct {
	log         *logger.Logger
	mu          sync.RWMutex
	subscribers map[string]map[string]chan []byte
}

func NewDispatcher(topics []models.Topic, log *logger.Logger) *Dispatcher {
	subs := make(map[string]map[string]chan []byte)

	for _, topic := range topics {
		subs[string(topic)] = make(map[string]chan []byte)
	}

	return &Dispatcher{
		log:         log,
		mu:          sync.RWMutex{},
		subscribers: subs,
	}
}

func (d *Dispatcher) RegisterUsingTopics(topics []models.Topic, client *models.WebsocketClient) error {
	topicsString := make([]string, 0)
	for _, topic := range topics {
		topicsString = append(topicsString, string(topic))
	}
	return d.Register(topicsString, client)
}

func (d *Dispatcher) Register(topics []string, client *models.WebsocketClient) error {
	if client == nil {
		return ErrClientRequired
	}
	if client.Id == "" {
		return ErrClientIdRequired
	}
	if client.Send == nil {
		return ErrClientSendChannelRequired
	}

	topics = handleTopics(topics)

	d.mu.Lock()
	defer d.mu.Unlock()
	for _, topic := range topics {
		clients, exists := d.subscribers[topic]
		if !exists {
			return ErrTopicNonExistent
		}
		if _, exists := clients[client.Id]; exists {
			d.log.Warn("Trying to Register client already subscribed", client.Id, topic)
			return nil
		}
		d.log.Debug("Registered client", client.Id, "to topic", topic)
		clients[client.Id] = client.Send
	}
	return nil
}

func (d *Dispatcher) Unregister(topic string, clientId string, sendCh chan []byte) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	clients, exists := d.subscribers[topic]
	if !exists {
		return ErrTopicNonExistent
	}
	current, exists := clients[clientId]
	if !exists || current != sendCh {
		return nil
	}
	delete(clients, clientId)
	return nil
}

func (d *Dispatcher) UnregisterFromAllTopics(clientId string, sendCh chan []byte) error {
	for topic := range d.subscribers {
		err := d.Unregister(topic, clientId, sendCh)
		if err != nil {
			d.log.Error("Error unregistering client", err)
		}
	}
	return nil
}

func (d *Dispatcher) Dispatch(topic string, message []byte) error {
	d.mu.RLock()
	clients, exists := d.subscribers[topic]
	d.mu.RUnlock()

	if !exists {
		return ErrTopicNonExistent
	}

	if len(clients) == 0 {
		d.log.Warn("Dispatch message of topic not listened by anyone", topic)
		return nil
	}

	var wg sync.WaitGroup
	for _, sendChan := range clients {
		wg.Add(1)
		go func(ch chan []byte) {
			defer wg.Done()
			select {
			case ch <- message:
			default:
				// Canal cheio - log ou ignorar
			}
		}(sendChan)
	}

	wg.Wait()
	return nil
}

func (d *Dispatcher) subscribeToTopic(ctx context.Context, broker broker.Broker, topic models.Topic) <-chan *models.Notification {
	const retryInterval = 5 * time.Second
	var notifications <-chan *models.Notification
	for {
		select {
		case <-ctx.Done():
			d.log.Warn("Context cancelled before subscribing to topic", topic)
			return nil
		default:
			var err error
			notifications, err = broker.SubscribeNotifications(ctx, topic)
			if err != nil {
				d.log.Error("Failed to subscribe dispatch to topic", topic, "retrying in", retryInterval, err)
				select {
				case <-time.After(retryInterval):
					continue
				case <-ctx.Done():
					d.log.Warn("Context cancelled before subscribing to topic", topic)
					return nil
				}
			}
			d.log.Info("Dispatch subscribed successfully to topic!", topic)
			return notifications
		}
	}
}

func (d *Dispatcher) ConsumeTopic(ctx context.Context, broker broker.Broker, topic models.Topic) {
	d.log.Info("Starting to consume messages for topic", topic)
	ch := d.subscribeToTopic(ctx, broker, topic)
	if ch == nil {
		return
	}
	d.log.Debug("Starting to dispatch messages from topic to subscribers!", topic)
	for notification := range ch {
		d.SendMessageToSubscribers(notification)
	}
	d.log.Debug("Stopped to dispatch messages from topic to subscribers!", topic)
}

func (d *Dispatcher) ConsumeMessages(ctx context.Context, broker broker.Broker) {
	d.log.Info("Starting to consume messages for all topics")
	for _, topic := range models.GetAllNotificationTopics() {
		go d.ConsumeTopic(ctx, broker, topic)
	}
	d.log.Info("Consume messages started")
}

func (d *Dispatcher) SendMessageToSubscribers(notification *models.Notification) {
	data, err := json.Marshal(notification)
	if err != nil {
		d.log.Error("failed to marshal notification while sending to subscribers", err)
		return
	}
	d.log.Debug("Notification arrived at dispatch", "topic", notification.Topic, "title", notification.Title, "clientId", notification.ClientID)
	switch notification.Topic {
	case models.ALL:
		for _, ch := range d.subscribers[string(models.ALL)] {
			ch <- data
		}
	case models.PRIVATE:
		if ch, exists := d.subscribers[string(models.PRIVATE)][notification.ClientID]; exists {
			ch <- data
		}
	case "":
		d.log.Error(ErrTopicRequired)
	default:
		if _, exists := d.subscribers[string(notification.Topic)]; !exists {
			d.log.Error(ErrTopicNonExistent, notification.Topic)
			return
		}
		for _, ch := range d.subscribers[string(notification.Topic)] {
			ch <- data
		}
	}
}

func handleTopics(topics []string) []string {
	// Filter topics
	filtered := make([]string, 0)
	allTopics := models.GetAllNotificationTopics()
	for _, topic := range topics {
		if slices.Contains(allTopics, models.Topic(topic)) &&
			topic != string(models.ALL) &&
			topic != string(models.PRIVATE) {
			filtered = append(filtered, topic)
		}
	}
	return append(
		filtered,
		string(models.ALL),
		string(models.PRIVATE),
	)
}

func (p *Dispatcher) Shutdown(ctx context.Context) {

}
