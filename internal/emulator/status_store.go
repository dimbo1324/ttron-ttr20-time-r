package emulator

func (s *Service) recordRequest() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.status.TotalRequests++
	s.status.LastRequestTime = s.now()
}

func (s *Service) recordResponse() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.status.TotalResponses++
	s.status.LastResponseTime = s.now()
}

func (s *Service) recordProtocolError(errText string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.status.TotalProtocolErrors++
	s.status.LastError = errText
}

func (s *Service) connectionOpened() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.status.ActiveConnections++
	s.status.TotalConnections++
}

func (s *Service) connectionClosed() {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.status.ActiveConnections > 0 {
		s.status.ActiveConnections--
	}
}

func (s *Service) setRunning(running bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.status.Running = running
}
