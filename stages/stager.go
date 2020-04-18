package stages

type Stager struct{ stages []Stage }

func NewStager(stages []Stage) *Stager { return &Stager{stages: stages} }

func (s *Stager) NextStage() Stage {
	if len(s.stages) > 1 {
		s.stages = s.stages[1:]
	}
	return s.stages[0]
}

func (s *Stager) Stage() Stage { return s.stages[0] }

func (s Stager) MarshalYAML() (interface{}, error) { return s.stages, nil }

func (s *Stager) UnmarshalYAML(unmarshal func(interface{}) error) error {
	return unmarshal(&s.stages)
}
