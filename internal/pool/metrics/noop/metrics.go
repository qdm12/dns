package noop

type Metrics struct{}

func New() (metrics *Metrics) {
	return &Metrics{}
}

func (m *Metrics) Init(string) {}

func (m *Metrics) LiveConnInc(string) {}

func (m *Metrics) DeadConnInc(string) {}

func (m *Metrics) RemovedConnsAdd(string, uint) {}

func (m *Metrics) GetConnInc(string, string) {}

func (m *Metrics) PutConnInc(string, string) {}

func (m *Metrics) NewConnsInc(string, string) {}

func (m *Metrics) RenewConnInc(string, string, string) {}
