package beater

import (
	"errors"
	"fmt"
	"github.com/elastic/beats/libbeat/beat"
	"github.com/elastic/beats/libbeat/common"
	"github.com/elastic/beats/libbeat/logp"
	"github.com/kckecheng/powermaxbeat/config"
	"github.com/kckecheng/storagemetric/dell/emc/powermax"
	"time"
)

// Powermaxbeat configuration.
type Powermaxbeat struct {
	done   chan struct{}
	config config.Config
	client beat.Client
	pmax   *powermax.PowerMax
}

// New creates an instance of powermaxbeat.
func New(b *beat.Beat, cfg *common.Config) (beat.Beater, error) {
	c := config.DefaultConfig
	if err := cfg.Unpack(&c); err != nil {
		return nil, fmt.Errorf("Error reading config file: %v", err)
	}

	bt := &Powermaxbeat{
		done:   make(chan struct{}),
		config: c,
	}

	host := bt.config.Host
	port := bt.config.Port
	username := bt.config.Username
	password := bt.config.Password
	symmid := bt.config.Symmid
	if host == "" || username == "" || password == "" || symmid == "" {
		return nil, errors.New("Unisphere address(IP/FQDN), username, password and PowerMax symmetric ID must be specified")
	}
	if port == "" {
		port = "8443"
	}
	pmax, err := powermax.New(host, port, username, password, symmid)
	if err != nil {
		return nil, err
	}
	bt.pmax = pmax

	interval := int(bt.config.Period / time.Second)
	// It is not recommended to query PowerMax performance through Unisphere too
	// frequently since all 0 will be returned under such a condition
	if interval < 30 {
		return nil, fmt.Errorf("The configured period %d must be larger than 30", interval)
	}
	return bt, nil
}

// Run starts powermaxbeat.
func (bt *Powermaxbeat) Run(b *beat.Beat) error {
	logp.Info("powermaxbeat is running! Hit CTRL-C to stop it.")

	var err error
	bt.client, err = b.Publisher.Connect()
	if err != nil {
		return err
	}

	interval := int(bt.config.Period / time.Second)
	ticker := time.NewTicker(bt.config.Period)
	for {
		select {
		case <-bt.done:
			return nil
		case <-ticker.C:
		}

		current_tm := time.Now()
		from_tm := current_tm.Add(-time.Second * time.Duration(interval))
		arrmetric := bt.pmax.GetArrayMetric(from_tm, current_tm)

		event := beat.Event{
			Timestamp: current_tm,
			Fields: common.MapStr{
				"type":                   "powermaxbeat",
				"unisphere":              bt.config.Host,
				"powermax":               bt.config.Symmid,
				"iops(host)":             arrmetric.HostIOs,
				"read(host)":             arrmetric.HostReads,
				"write(host)":            arrmetric.HostWrites,
				"throughput(host in MB)": arrmetric.HostMBReads + arrmetric.HostMBWritten,
				"iops":                   arrmetric.FEReadReqs + arrmetric.FEWriteReqs,
				"iops(read)":             arrmetric.FEReadReqs,
				"iops(write)":            arrmetric.FEWriteReqs,
				"rst(read)":              arrmetric.ReadResponseTime,
				"rst(write)":             arrmetric.WriteResponseTime,
				"utilization(FE)":        arrmetric.FEUtilization,
			},
		}
		bt.client.Publish(event)
		logp.Info("Event sent")
	}
}

// Stop stops powermaxbeat.
func (bt *Powermaxbeat) Stop() {
	bt.client.Close()
	close(bt.done)
}
