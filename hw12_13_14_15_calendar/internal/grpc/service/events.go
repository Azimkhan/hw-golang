package service

import (
	"context"
	"errors"

	"github.com/Azimkhan/hw-golang/hw12_13_14_15_calendar/gen/events/pb"
	"github.com/Azimkhan/hw-golang/hw12_13_14_15_calendar/internal/app"
	"github.com/Azimkhan/hw-golang/hw12_13_14_15_calendar/internal/storage/model"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type EventsService struct {
	app *app.App
	pb.UnimplementedEventServiceServer
}

func NewEventsService(app *app.App) *EventsService {
	return &EventsService{app: app}
}

func (s *EventsService) CreateEvent(ctx context.Context, r *pb.CreateEventRequest) (
	*pb.CreateEventResponse, error,
) {
	event := s.grpcToInternal(r.Event)
	if err := s.app.Storage.CreateEvent(ctx, event); err != nil {
		return nil, err
	}
	return &pb.CreateEventResponse{
		Event: s.internalToGrpc(event),
	}, nil
}

func (s *EventsService) UpdateEvent(ctx context.Context, r *pb.UpdateEventRequest) (
	*pb.UpdateEventResponse, error,
) {
	event := s.grpcToInternal(r.Event)
	if err := s.app.Storage.UpdateEvent(ctx, event); err != nil {
		return nil, err
	}
	return &pb.UpdateEventResponse{
		Event: s.internalToGrpc(event),
	}, nil
}

func (s *EventsService) RemoveEvent(ctx context.Context, r *pb.RemoveEventRequest) (
	*pb.RemoveEventResponse, error,
) {
	if err := s.app.Storage.RemoveEvent(ctx, r.GetId()); err != nil {
		return nil, err
	}
	return &pb.RemoveEventResponse{}, nil
}

func (s *EventsService) FilterEventsByDay(ctx context.Context, r *pb.FilterEventsByDayRequest) (
	*pb.FilterEventsByDayResponse, error,
) {
	if r.GetDate() == nil {
		return nil, errors.New("date is not specified")
	}
	events, err := s.app.Storage.FilterEventsByDay(ctx, r.GetDate().AsTime())
	if err != nil {
		return nil, err
	}
	return &pb.FilterEventsByDayResponse{
		Events: s.internalSliceToGrpc(events),
	}, nil
}

func (s *EventsService) FilterEventsByWeek(ctx context.Context, r *pb.FilterEventsByWeekRequest) (
	*pb.FilterEventsByWeekResponse, error,
) {
	if r.GetDate() == nil {
		return nil, errors.New("date is not specified")
	}
	events, err := s.app.Storage.FilterEventsByWeek(ctx, r.GetDate().AsTime())
	if err != nil {
		return nil, err
	}
	return &pb.FilterEventsByWeekResponse{
		Events: s.internalSliceToGrpc(events),
	}, nil
}

func (s *EventsService) FilterEventsByMonth(ctx context.Context, r *pb.FilterEventsByMonthRequest) (
	*pb.FilterEventsByMonthResponse, error,
) {
	if r.GetDate() == nil {
		return nil, errors.New("date is not specified")
	}
	events, err := s.app.Storage.FilterEventsByMonth(ctx, r.GetDate().AsTime())
	if err != nil {
		return nil, err
	}
	return &pb.FilterEventsByMonthResponse{
		Events: s.internalSliceToGrpc(events),
	}, nil
}

func (s *EventsService) internalSliceToGrpc(events []*model.Event) []*pb.Event {
	res := make([]*pb.Event, len(events))
	for i, e := range events {
		res[i] = s.internalToGrpc(e)
	}
	return res
}

func (s *EventsService) grpcToInternal(g *pb.Event) *model.Event {
	return &model.Event{
		ID:          g.GetId(),
		Title:       g.GetTitle(),
		StartTime:   g.GetStart().AsTime(),
		EndTime:     g.GetEnd().AsTime(),
		UserID:      g.GetUserId(),
		NotifyDelta: int(g.GetNotifyDelta()),
	}
}

func (s *EventsService) internalToGrpc(e *model.Event) *pb.Event {
	return &pb.Event{
		Id:          e.ID,
		Title:       e.Title,
		Start:       timestamppb.New(e.StartTime),
		End:         timestamppb.New(e.EndTime),
		UserId:      e.UserID,
		NotifyDelta: uint32(e.NotifyDelta),
	}
}
