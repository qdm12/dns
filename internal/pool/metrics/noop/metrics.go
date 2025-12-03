package noop

type Metrics struct{}

func New() (metrics *Metrics) {
	return &Metrics{}
}

func (m *Metrics) ConnsAdd(string, int)      {}
func (m *Metrics) LiveConnsAdd(string, int)  {}
func (m *Metrics) InUseConnsAdd(string, int) {}
func (m *Metrics) RenewRequestsInc(string)   {}
func (m *Metrics) RenewalsInc(string)        {}
