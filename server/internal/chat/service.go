package chat

import (
	"context"
	"sync"
	"time"

	"github.com/google/uuid"
	pb "github.com/you/chatpoc/server/api"
)

type event = pb.Event

type client struct {
	send chan *event
}

// room giữ history ngắn để demo (in-memory)
type room struct {
	id      string
	seq     int64
	history []*event // giới hạn N
	clients map[*client]bool
	mu      sync.RWMutex
}

type Hub struct {
	rooms map[string]*room
	mu    sync.RWMutex
}

func NewHub() *Hub { return &Hub{rooms: make(map[string]*room)} }

func (h *Hub) getRoom(id string) *room {
	h.mu.Lock()
	defer h.mu.Unlock()
	if r, ok := h.rooms[id]; ok {
		return r
	}
	r := &room{id: id, clients: make(map[*client]bool)}
	h.rooms[id] = r
	return r
}

func (r *room) nextSeq() int64 {
	r.seq++
	return r.seq
}

func (r *room) appendHistory(ev *event) {
	const maxHistory = 500 // đủ cho PoC
	if len(r.history) >= maxHistory {
		r.history = r.history[1:]
	}
	r.history = append(r.history, ev)
}

func (r *room) broadcast(ev *event) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for c := range r.clients {
		select {
		case c.send <- ev:
		default:
			// client chậm, bỏ qua để không kẹt server
		}
	}
}

type Service struct {
	pb.UnimplementedChatServiceServer
	hub *Hub
}

func NewService(h *Hub) *Service { return &Service{hub: h} }

func (s *Service) Join(req *pb.JoinRequest, stream pb.ChatService_JoinServer) error {
	r := s.hub.getRoom(req.RoomId)
	c := &client{send: make(chan *event, 128)}

	// gửi history theo since_seq
	r.mu.Lock()
	r.clients[c] = true
	var snapshot []*event
	for _, ev := range r.history {
		if ev.Seq > req.SinceSeq {
			snapshot = append(snapshot, ev)
		}
	}
	r.mu.Unlock()

	// đẩy history trước
	for _, ev := range snapshot {
		if err := stream.Send(ev); err != nil {
			r.mu.Lock()
			delete(r.clients, c)
			r.mu.Unlock()
			return err
		}
	}

	// sau đó lắng nghe realtime
	ctx := stream.Context()
	for {
		select {
		case <-ctx.Done():
			r.mu.Lock()
			delete(r.clients, c)
			r.mu.Unlock()
			return nil
		case ev := <-c.send:
			if err := stream.Send(ev); err != nil {
				r.mu.Lock()
				delete(r.clients, c)
				r.mu.Unlock()
				return err
			}
		}
	}
}

func (s *Service) Send(ctx context.Context, in *pb.SendRequest) (*pb.Event, error) {
	r := s.hub.getRoom(in.RoomId)
	r.mu.Lock()
	ev := &pb.Event{
		RoomId:      in.RoomId,
		Id:          uuid.New().String(),
		Seq:         r.nextSeq(),
		Type:        in.Type,
		UserId:      in.UserId,
		Text:        in.Text,
		PayloadJson: in.PayloadJson,
		CreatedAtMs: time.Now().UnixMilli(),
	}
	r.appendHistory(ev)
	r.mu.Unlock()
	r.broadcast(ev)
	return ev, nil
}
