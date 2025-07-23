package main

import (
	math "math/rand/v2"
	pb "ride-sharing/shared/proto/driver"
	"ride-sharing/shared/util"
	"sync"

	"github.com/mmcloughlin/geohash"
)

type Service struct {
	drivers []*driverInMap
	mu      sync.Mutex
}

type driverInMap struct {
	Driver *pb.Driver
}

func NewService() *Service {
	return &Service{
		drivers: make([]*driverInMap, 0),
	}
}

func (s *Service) RegisterDriver(driverId string, packageSlug string) (*pb.Driver, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	randomIndex := math.IntN(len(PredefinedRoutes))
	randomRoute := PredefinedRoutes[randomIndex]

	// we can ignore this property for now, but it must be sent to the frontend.
	geohash := geohash.Encode(randomRoute[0][0], randomRoute[0][1])

	randomPlate := GenerateRandomPlate()
	randomAvatar := util.GetRandomAvatar(1)

	driver := &pb.Driver{
		Id:             driverId,
		Geohash:        geohash,
		Location:       &pb.Location{Latitude: randomRoute[0][0], Longitude: randomRoute[0][1]},
		Name:           "Lando Norris",
		PackageSlug:    packageSlug,
		ProfilePicture: randomAvatar,
		CarPlate:       randomPlate,
	}

	s.drivers = append(s.drivers, &driverInMap{Driver: driver})
	return driver, nil
}

func (s *Service) UnregisterDriver(driverId string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i, driver := range s.drivers {
		if driver.Driver.Id == driverId {
			s.drivers = append(s.drivers[:i], s.drivers[i+1:]...)
			break
		}
	}
}

/*
append(s.drivers[:i], s.drivers[i+1:]...)

we want to remove the item C

[A,B,C,D,E,F]

Step 1 - Get all of the elements before C
s.drivers[:i] // [A,B]

Step 2 - Get all of the elements after C
s.drivers[i+1:] // [D,E,F]

Step 3 - Append the two slices together
 append(s.drivers[:i], s.drivers[i+1:]...)
[A,B] + [D,E,F] = [A,B,D,E,F]
*/
