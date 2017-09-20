package consul

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestConsulClient(t *testing.T) {
	Convey("test consult client", t, func() {
		cc := NewConsulClient("http://192.168.150.220/sd", nil)
		name := "consult-test"
		address := "192.168.150.220"
		port := 4400

		Convey("register service should be ok", func() {
			err := cc.RegisterService(&AgentServiceRegistration{
				Name:    name,
				Tags:    []string{"mingjun.zhou", "golang"},
				Address: address,
				Port:    port,
				Check: &AgentServiceCheck{
					HTTP:     "http://192.168.150.220:4400",
					Method:   "GET",
					Interval: "10s",
					Timeout:  "1s",
				},
			})
			So(err, ShouldBeNil)

			Convey("list services should return the registered service", func() {
				services, err := cc.ListServices()
				So(err, ShouldBeNil)
				service, ok := services[name]
				So(ok, ShouldBeTrue)
				So(service.Address, ShouldEqual, address)
				So(service.Port, ShouldEqual, port)
			})
		})

		Convey("deregister service should be ok", func() {
			err := cc.DeregisterService(name)
			So(err, ShouldBeNil)
			Convey("list services should not return the registered service", func() {
				services, err := cc.ListServices()
				So(err, ShouldBeNil)
				_, ok := services[name]
				So(ok, ShouldBeFalse)
			})
		})
	})
}
