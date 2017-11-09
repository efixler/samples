package publish 

import (
	"cloud.google.com/go/pubsub"
	"context"
	"errors"
 	"google.golang.org/appengine/log"
	"github.com/taxat/api/config"
)

var (
	UnexpectedClientFetcher = errors.New("Client fetcher not of the expected type")
)

type Mapify interface {
	ToMap() map[string]string
}

func event(ctx context.Context, topic string, message pubsub.Message) {
	topic 	= config.Conf().PubSubRootNode() + "." + topic
	c, err 	:= client(ctx) 
	if err != nil {
		log.Errorf(ctx, "Error getting pubsub client (aborting): %s", err)
		return
	}
	t := c.Topic(topic) //instantiating t is somewhat expensive, would be better to reuse this
	defer t.Stop() 
	r := t.Publish(ctx, &message)
    go func() { 
    	id, err := r.Get(ctx)
    	if err != nil {
        	log.Errorf(ctx, "Error publishing message to %s topic: %s", topic, err)
        	return
    	} else {
    		log.Debugf(ctx, "Published message (ID[%s]) to %s", id, topic)
    	}
    }()  
}

func Event(ctx context.Context, topic string, message string, attributes map[string]string) {
	//go 
	event(ctx, topic, pubsub.Message{Data: []byte(message), Attributes: attributes})
	log.Debugf(ctx, "Fired event %s to %s", message, topic)
}

type pubsubClientContextKey string 

func client(ctx context.Context) (*pubsub.Client, error) {
	v := ctx.Value(pubsubClientContextKey("PSCF"))
	switch v.(type) {
		case func()(*pubsub.Client, error): 
			return v.(func()(*pubsub.Client, error))()
		default:
			return nil, UnexpectedClientFetcher
	}
}

func clientFetcher(ctx context.Context) func() (*pubsub.Client, error) {
	psp := config.Conf().PubSubProjectID()
	var client *pubsub.Client	
	f := func() (*pubsub.Client, error) {
		if client == nil {
			if c, err := pubsub.NewClient(ctx, psp); err != nil {
				return nil, err
			} else {
				client = c
			}
		} 
		return client, nil
	}
	return f
}

func AttachPublishClient(ctx context.Context) context.Context {
	ctx = context.WithValue(ctx, pubsubClientContextKey("PSCF"), clientFetcher(ctx))
	return ctx
}
