package oracleInterface

import (
	"github.com/robfig/cron/v3"
	"sync"
)

type Cron struct {
	sync.Mutex
	cr        *cron.Cron
	crRunning map[int]string
}

func newCron(opts ...cron.Option) *Cron {
	return &Cron{
		cr:        cron.New(opts...),
		crRunning: map[int]string{},
	}
}

func (c *Cron) Start() {
	c.Lock()
	defer c.Unlock()
	c.cr.Start()
}

func (c *Cron) Stop() {
	c.Lock()
	defer c.Unlock()
	c.cr.Stop()
}

func (c *Cron) Remove(ids ...int) {
	c.Lock()
	defer c.Unlock()
	for _, id := range ids {
		c.cr.Remove(cron.EntryID(id))
		delete(c.crRunning, id)
	}
}

func (c *Cron) AddFunc(spec string, cmd func()) (int, error) {
	c.Lock()
	defer c.Unlock()
	id, err := c.cr.AddFunc(spec, cmd)
	if err != nil {
		return 0, err
	}
	c.crRunning[int(id)] = spec
	return int(id), nil
}
func (c *Cron) Query(id int) string {
	return c.crRunning[id]
}
