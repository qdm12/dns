package noop

type Metrics struct{}

func New() (metrics *Metrics) {
	return &Metrics{}
}

func (m *Metrics) NewConnsInc(_, _ string)          {}
func (m *Metrics) RenewedConnsInc(_, _, _ string)   {}
func (m *Metrics) DeadConnsInc(_ string)            {}
func (m *Metrics) RemovedConnsAdd(_ string, _ uint) {}
func (m *Metrics) GetConnsInc(_, _ string)          {}
func (m *Metrics) PutConnsInc(_, _ string)          {}
