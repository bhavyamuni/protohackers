package speeddaemon

import (
	"log"
	"math"
	"net"
)

type IAmCameraMessage struct {
	MessageType
	Road  uint16
	Mile  uint16
	Limit uint16
}

type Camera struct {
	Road  uint16
	Mile  uint16
	Limit uint16
	Conn  *net.Conn
}

type PlateMessage struct {
	MessageType
	Plate     string
	Timestamp uint32
}

type Observation struct {
	Plate     string
	Timestamp uint32
	Camera    Camera
}

func (m *PlateMessage) Handle(s *SpeedDaemonServer, conn *net.Conn) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.cameras[conn]; !ok {
		s.SendError(conn, "Camera not registered")
		return
	}

	if _, ok := s.observations[m.Plate]; !ok {
		s.observations[m.Plate] = []Observation{}
	}

	newObservation := Observation{
		Plate:     m.Plate,
		Timestamp: m.Timestamp,
		Camera:    s.cameras[conn],
	}

	go CheckSpeedViolation(s.observations[m.Plate], newObservation, s.tickets[newObservation.Camera.Road])
	s.observations[m.Plate] = append(s.observations[m.Plate], newObservation)
}

func (m *IAmCameraMessage) Handle(s *SpeedDaemonServer, conn *net.Conn) {
	s.mu.Lock()
	defer s.mu.Unlock()
	log.Println("IAmCameraMessage received: ", (*conn).RemoteAddr())
	if _, ok := s.dispatchers[conn]; ok {
		s.SendError(conn, "Dispatcher already registered")
		return
	}
	if _, ok := s.cameras[conn]; ok {
		s.SendError(conn, "Camera already registered")
		return
	}
	s.cameras[conn] = Camera{Road: m.Road, Mile: m.Mile, Limit: m.Limit, Conn: conn}
	if _, ok := s.tickets[m.Road]; !ok {
		s.tickets[m.Road] = make(chan *Ticket)
	}
}

func CreateTicket(observation1 Observation, observation2 Observation, speed float64) Ticket {
	if observation1.Timestamp > observation2.Timestamp {
		observation1, observation2 = observation2, observation1
	}
	ticket := Ticket{
		Plate:      observation1.Plate,
		Road:       observation1.Camera.Road,
		Mile1:      observation1.Camera.Mile,
		Timestamp1: observation1.Timestamp,
		Mile2:      observation2.Camera.Mile,
		Timestamp2: observation2.Timestamp,
		Speed:      uint16(speed),
	}

	return ticket
}

func CheckSpeedViolation(observations []Observation, currentObservation Observation, tickets chan<- *Ticket) {
	for i := range observations {
		if observations[i].Camera.Road == currentObservation.Camera.Road {
			// Skip if observations are more than 1 day apart (86400 seconds)
			if math.Abs(float64(observations[i].Timestamp)-float64(currentObservation.Timestamp)) > 86400 {
				continue
			}

			speed := FindSpeed(observations[i].Camera.Mile, currentObservation.Camera.Mile, observations[i].Timestamp, currentObservation.Timestamp)
			if math.Abs(speed) >= (float64(currentObservation.Camera.Limit) + 0.5) {
				ticket := CreateTicket(observations[i], currentObservation, math.Abs(speed)*100)
				tickets <- &ticket
			}
		}
	}
}

func FindSpeed(distance1 uint16, distance2 uint16, time1 uint32, time2 uint32) float64 {
	distance := float64(distance2) - float64(distance1)
	time := (float64(time2) - float64(time1)) / 3600
	return distance / time
}
