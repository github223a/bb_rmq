package rmq

import uuid "github.com/satori/go.uuid"

type ChannelSettings struct {
	ExchangeName    string         `json:"exchangeName,omitempty"`
	ExchangeType    string         `json:"exchangeType,omitempty"`
	QueueName       string         `json:"queueName,omitempty"`
	QueueOptions    *QueueSettings `json:"queueOptions,omitempty"`
	ConsumeActivate bool           `json:"consumeActivate,omitempty"`
	BindingKey      string         `json:"bindingKey,omitempty"`
}

type QueueSettings struct {
	MessageTTL int32 `json:"messageTtl,omitempty"`
	Durable    bool  `json:"durable,omitempty"`
	NoAck      bool  `json:"noAck,omitempty"`
	AutoDelete bool  `json:"autoDelete,omitempty"`
}

type Request struct {
	ID            uuid.UUID              `json:"id"`
	Namespace     string                 `json:"namespace"`
	Method        string                 `json:"method"`
	Domain        *string                `json:"domain"`
	Locale        *string                `json:"locale"`
	Params        map[string]interface{} `json:"params"`
	ResponseQueue string                 `json:"responseQueue"`
	Source        string                 `json:"source,omitempty"`
	Subscribe     bool                   `json:"subscribe,omitempty"`
	CacheKey      string                 `json:"cacheKey,omitempty"`
	Token         string                 `json:"token,omitempty"`
}

type SuccessResponse struct {
	ID            uuid.UUID              `json:"id"`
	Namespace     string                 `json:"namespace"`
	Method        string                 `json:"method"`
	Domain        string                 `json:"domain"`
	Locale        string                 `json:"locale"`
	Result        map[string]interface{} `json:"result"`
	Source        string                 `json:"source"`
	ResponseQueue string                 `json:"responseQueue"`
	CacheKey      string                 `json:"cacheKey,omitempty"`
	Token         string                 `json:"token,omitempty"`
}

type ErrorResponse struct {
	ID            uuid.UUID              `json:"id"`
	Namespace     string                 `json:"namespace"`
	Method        string                 `json:"method"`
	Domain        string                 `json:"domain"`
	Locale        string                 `json:"locale"`
	Error         map[string]interface{} `json:"error"`
	Source        string                 `json:"source"`
	ResponseQueue string                 `json:"responseQueue"`
	CacheKey      string                 `json:"cacheKey,omitempty"`
	Token         string                 `json:"token,omitempty"`
}
