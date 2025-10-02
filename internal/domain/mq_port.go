package domain

type Message struct {
	Key 	[]byte
	Value 	[]byte
}

type PublisherPort interface {
	Publish(topic string, msgs ...Message) error
}

type SubscriberPort interface {
	Subscribe(topic, groupID string) (<-chan Message, error)
}