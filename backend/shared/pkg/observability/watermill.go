package observability

import (
	"context"

	"github.com/ThreeDotsLabs/watermill/message"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

const (
	WatermillTracerName = "watermill-otel"
)

// TracingPublisher decorates a Watermill Publisher to add OTel tracing
type TracingPublisher struct {
	publisher message.Publisher
	tracer    trace.Tracer
}

func NewTracingPublisher(pub message.Publisher) message.Publisher {
	return &TracingPublisher{
		publisher: pub,
		tracer:    otel.Tracer(WatermillTracerName),
	}
}

func (p *TracingPublisher) Publish(topic string, messages ...*message.Message) error {
	for _, msg := range messages {
		// Create span
		ctx := msg.Context()
		ctx, span := p.tracer.Start(ctx, "watermill.publish",
			trace.WithAttributes(
				attribute.String("messaging.system", "kafka"), // Assuming Kafka
				attribute.String("messaging.destination", topic),
				attribute.String("messaging.message_id", msg.UUID),
			),
			trace.WithSpanKind(trace.SpanKindProducer),
		)
		defer span.End()

		// Inject context into message metadata
		otel.GetTextMapPropagator().Inject(ctx, propagation.MapCarrier(msg.Metadata))

		// Update message context
		msg.SetContext(ctx)
	}

	err := p.publisher.Publish(topic, messages...)
	if err != nil {
		// We can't record error on a specific span easily here because we have multiple messages
		// But usually Publish is atomic or batch.
		// For simplicity, we just return error.
		// Ideally we should link errors.
	}
	return err
}

func (p *TracingPublisher) Close() error {
	return p.publisher.Close()
}

// TracingSubscriber decorates a Watermill Subscriber to add OTel tracing
type TracingSubscriber struct {
	subscriber message.Subscriber
	tracer     trace.Tracer
}

func NewTracingSubscriber(sub message.Subscriber) message.Subscriber {
	return &TracingSubscriber{
		subscriber: sub,
		tracer:     otel.Tracer(WatermillTracerName),
	}
}

func (s *TracingSubscriber) Subscribe(ctx context.Context, topic string) (<-chan *message.Message, error) {
	messages, err := s.subscriber.Subscribe(ctx, topic)
	if err != nil {
		return nil, err
	}

	out := make(chan *message.Message)

	go func() {
		defer close(out)
		for msg := range messages {
			// Extract context from metadata
			propagator := otel.GetTextMapPropagator()
			ctx := propagator.Extract(msg.Context(), propagation.MapCarrier(msg.Metadata))

			// Start consumer span
			ctx, span := s.tracer.Start(ctx, "watermill.consume",
				trace.WithAttributes(
					attribute.String("messaging.system", "kafka"),
					attribute.String("messaging.destination", topic),
					attribute.String("messaging.message_id", msg.UUID),
					attribute.String("messaging.operation", "receive"),
				),
				trace.WithSpanKind(trace.SpanKindConsumer),
			)

			// Update message context so handler uses this span
			msg.SetContext(ctx)

			// Ensure span ends only after message is processed (Ack/Nack)
			go func() {
				select {
				case <-msg.Acked():
				case <-msg.Nacked():
				}
				span.End()
			}()

			out <- msg
		}
	}()

	return out, nil
}

func (s *TracingSubscriber) Close() error {
	return s.subscriber.Close()
}
