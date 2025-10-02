//
// SPDX-License-Identifier: BSD-3-Clause
//

package smc

import (
	"encoding/json"
	"errors"

	"github.com/myelintek/gofish/common"
	"github.com/myelintek/gofish/redfish"
)

// ErrActionNotSupported is returned when the requested OEM-specific action
// does not appear to be supported. This might happen when a device is new
// or upgraded to a new firmware that follows the DMTF standards.
var ErrActionNotSupported = errors.New("oem-specific action unsupported")

// Drive extends a redfish.Drive for additional OEM fields
type Drive struct {
	redfish.Drive

	// Fields from the SMC OEM section
	Temperature             int
	PercentageDriveLifeUsed int
	DriveFunctional         bool

	// indicateTarget is the uri to hit to change the light state
	indicateTarget string
}

// FromDrive returns an OEM-extended redfish drive
func FromDrive(drive *redfish.Drive) (Drive, error) {
	smcDrive := Drive{
		Drive: *drive,
	}

	var t struct {
		Oem struct {
			Supermicro struct {
				Temperature             int
				PercentageDriveLifeUsed int
				DriveFunctional         bool
			} `json:"Supermicro"`
		} `json:"Oem"`
		Actions struct {
			Oem struct {
				DriveIndicate    common.ActionTarget `json:"#Drive.Indicate"`
				SmcDriveIndicate common.ActionTarget `json:"#SmcDrive.Indicate"`
			} `json:"Oem"`
		} `json:"Actions"`
	}

	// Populate the Oem data
	if err := json.Unmarshal(drive.RawData, &t); err != nil {
		return smcDrive, err
	}

	smcDrive.Temperature = t.Oem.Supermicro.Temperature
	smcDrive.PercentageDriveLifeUsed = t.Oem.Supermicro.PercentageDriveLifeUsed
	smcDrive.DriveFunctional = t.Oem.Supermicro.DriveFunctional

	// We check both the SmcDriveIndicate and the DriveIndicate targets
	// in the Oem sections - certain models and bmc firmwares will mix
	// these up, so we check both
	smcDrive.indicateTarget = t.Actions.Oem.DriveIndicate.Target
	if t.Actions.Oem.SmcDriveIndicate.Target != "" {
		smcDrive.indicateTarget = t.Actions.Oem.SmcDriveIndicate.Target
	}

	return smcDrive, nil
}

// Indicate will set the indicator light activity, true for on, false for off
func (d *Drive) Indicate(active bool) error {
	// Return a common error to let the user try falling back on the normal gofish path
	if d.indicateTarget == "" {
		return ErrActionNotSupported
	}

	return d.Post(d.indicateTarget, map[string]interface{}{"Active": active})
}
