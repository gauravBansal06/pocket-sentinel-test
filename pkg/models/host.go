package models

import (
	"errors"
	"fmt"

	pb "github.com/LambdatestIncPrivate/protobuf/golang/bookkeeping/host/v1"
	v1 "github.com/LambdatestIncPrivate/protobuf/golang/core/v1"
	"github.com/golang/protobuf/ptypes"
	"github.com/google/uuid"
)

type Host struct {
	// embed common model
	*CommonModel
	PrivateIP         string `db:"private_ip" json:"privateIP"`
	PublicIP          string `db:"public_ip"`
	Region            int
	Provider          int
	Status            int
	Hash              string
	Fqdn              string
	RAM               int
	ClockSpeed        int `db:"clock_speed"`
	Cores             int
	Storage           int
	CPULimit          int `db:"cpu_limit"`
	RAMLimit          int `db:"ram_limit"`
	StorageLimit      int `db:"storage_limit"`
	Capacity          int
	AvailableCapacity int    `db:"available_capacity"`
	OrganizationID    string `db:"organization_id"`
	Win10             bool   `db:"win_10"`
	Win81             bool   `db:"win_8_1"`
	Win8              bool   `db:"win_8"`
	Win7              bool   `db:"win_7"`
	WinXP             bool   `db:"win_xp"`
	MacBigsur         bool   `db:"mac_bigsur"`
	MacCatalina       bool   `db:"mac_catalina"`
	MacMojave         bool   `db:"mac_mojave"`
	MacHighSierra     bool   `db:"mac_high_sierra"`
	MacSierra         bool   `db:"mac_sierra"`
	MacElCapitan      bool   `db:"mac_el_capitan"`
	MacYosemite       bool   `db:"mac_yosemite"`
	MacMaverics       bool   `db:"mac_mavericks"`
	MacMountainLion   bool   `db:"mac_mountain_lion"`
	MacLion           bool   `db:"mac_lion"`
	MobileAndroid10   bool   `db:"mobile_android_10"`
	MobileAndroid9    bool   `db:"mobile_android_9"`
	MobileAndroid8    bool   `db:"mobile_android_8"`
	MobileAndroid7    bool   `db:"mobile_android_7"`
}

//Validate validates Host object for valid hash and id
func (h *Host) Validate() error {
	if h.Hash == "" || len(h.Hash) != 15 {
		return errors.New("Invalid hash")
	}

	// return error if id is not a valid uuid
	if _, err := uuid.Parse(h.ID); err != nil {
		return errors.New("Invalid ID. ID must a valid UUID")
	}

	return nil
}

// ToDto converts model host in grpc host
func (h *Host) ToDto() (*pb.Host, error) {
	// convert timestamps
	createdAt, err := ptypes.TimestampProto(h.Created)
	if err != nil {
		return nil, fmt.Errorf("Unable to convert created timestamp to protobuf format %s", err)
	}

	updatedAt, err := ptypes.TimestampProto(h.Updated)
	if err != nil {
		return nil, fmt.Errorf("Unable to convert updated timestamp to protobuf format :%s", err)
	}

	dto := pb.Host{
		Uuid:      h.ID,
		PrivateIp: h.PrivateIP,
		PublicIp:  h.PublicIP,
		Hash:      h.Hash,
		Region:    v1.Region(h.Region),
		Provider:  pb.Provider(h.Provider),
		Status:    pb.HostStatus(h.Status),
		Created:   createdAt,
		Updated:   updatedAt,
	}

	return &dto, nil
}

// HostFromDto converts grpc host struct to Host model
func HostFromDto(dto *pb.Host) (*Host, error) {
	host := Host{
		PrivateIP:   dto.PrivateIp,
		PublicIP:    dto.PublicIp,
		Hash:        dto.Hash,
		Region:      int(dto.Region),
		Provider:    int(dto.Provider),
		Status:      int(dto.Status),
		CommonModel: &CommonModel{},
	}

	host.ID = dto.Uuid

	// convert timestamps
	createdAt, err := ptypes.Timestamp(dto.Created)
	if err == nil {
		host.Created = createdAt
	}

	updatedAt, err := ptypes.Timestamp(dto.Updated)
	if err == nil {
		host.Updated = updatedAt
	}

	return &host, nil
}

func (h *Host) GetTableName() string {
	return "hosts"
}
