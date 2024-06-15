package serve

import (
	"context"
	"net"
	"strconv"

	"github.com/rs/zerolog/log"

	"github.com/morphy76/g-fe-server/internal/options"
)

func StartRouteRegistry(servOptions options.ServeOptions, incomingMessage chan []byte) (start func(), stop func()) {

	var udpConnection *net.UDPConn
	usePort, err := strconv.Atoi(servOptions.AnnouncePort)
	if err != nil {
		panic(err)
	}

	ctx, cancel := context.WithCancel(context.Background())

	start = func() {
		log.Info().Msg("Starting route registry")

		udpConnection, err = net.ListenUDP("udp", &net.UDPAddr{
			Port: usePort,
		})
		if err != nil {
			panic(err)
		}

		go func() {
			for {
				select {
				case <-ctx.Done():
					return
				default:
				}

				buffer := make([]byte, 1024)
				n, _, err := udpConnection.ReadFromUDP(buffer)
				if err != nil {
					if err, ok := err.(*net.OpError); ok && err.Err.Error() == "use of closed network connection" {
						return
					}
					log.Warn().Err(err).Msg("Error reading from UDP connection")
				}
				incomingMessage <- buffer[:n]
			}
		}()
	}

	stop = func() {
		log.Info().Msg("Stopping route registry")

		cancel()

		if udpConnection != nil {
			udpConnection.Close()
		}
	}

	return
}

func RegisterRoute(serveOptions options.ServeOptions, routeUri string) {
	dispatchUDP(serveOptions, routeUri)
}

func UnRegisterRoute(serveOptions options.ServeOptions, routeUri string) {
	dispatchUDP(serveOptions, routeUri)
}

func dispatchUDP(serveOptions options.ServeOptions, routeUri string) {

	usePort, err := strconv.Atoi(serveOptions.AnnouncePort)
	if err != nil {
		panic(err)
	}

	ips, err := net.LookupIP(serveOptions.AnnounceHost)
	if err != nil {
		panic(err)
	}

	local, err := net.ResolveUDPAddr("udp4", ":0")
	if err != nil {
		panic(err)
	}

	connections := make([]*net.UDPConn, 0)
	defer func() {
		for _, conn := range connections {
			if conn != nil {
				conn.Close()
			}
		}
	}()

	for _, ip := range ips {
		if ip.To4() == nil {
			continue
		}
		udpConnection, err := net.DialUDP("udp4", local, &net.UDPAddr{
			IP:   ip,
			Port: usePort,
		})
		if err != nil {
			log.Warn().Err(err).Msg("Error dialing UDP connection")
		}
		connections = append(connections, udpConnection)
	}

	for _, conn := range connections {
		if conn != nil {
			log.Trace().
				Str("route", routeUri).
				Msg("Announcing remote route")
			conn.Write([]byte(routeUri))
		}
	}
}
