package stages

// Stager is a stage sequence representation
type Stager struct{ stages []Stage }

// NewStager returns a new *Stager
func NewStager(stages ...Stage) *Stager { return &Stager{stages: stages} }

// NextStage advances to the next stage
func (s *Stager) NextStage() Stage {
	if len(s.stages) > 1 {
		s.stages = s.stages[1:]
	}
	return s.stages[0]
}

// Stage returns the current stage
func (s *Stager) Stage() Stage { return s.stages[0] }

// Append appends stages
func (s *Stager) Append(st ...Stage) { s.stages = append(s.stages, st...) }

// Stages returns the remaining stages
func (s *Stager) Stages() []Stage { return s.stages }

// MarshalYAML implement yaml.Marshaler
func (s Stager) MarshalYAML() (interface{}, error) { return s.stages, nil }

// UnmarshalYAML implement yaml.Unmarshaler
func (s *Stager) UnmarshalYAML(unmarshal func(interface{}) error) error {
	r := []Stage{}
	if err := unmarshal(&r); err != nil {
		return err
	}
	s.stages = r
	return nil
}
