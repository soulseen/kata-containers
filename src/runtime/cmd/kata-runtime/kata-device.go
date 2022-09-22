// Copyright (c) 2022 NetEase Inc.
//
// SPDX-License-Identifier: Apache-2.0
//

package main

import (
	"encoding/json"
	"fmt"
	containerdshim "github.com/kata-containers/kata-containers/src/runtime/pkg/containerd-shim-v2"
	"github.com/kata-containers/kata-containers/src/runtime/pkg/katautils"
	"github.com/kata-containers/kata-containers/src/runtime/pkg/utils/shimclient"
	"github.com/urfave/cli"
)

var (
	devPath string
)

var deviceSubCmds = []cli.Command{
	listDeviceCommand,
	attachDeviceCommand,
	detachDeviceCOmmand,
}

var kataDeviceCommand = cli.Command{
	Name:        "device",
	Usage:       "attach ot detach device to kata container",
	Subcommands: deviceSubCmds,
	Action: func(context *cli.Context) {
		cli.ShowSubcommandHelp(context)
	},
}

var listDeviceCommand = cli.Command{
	Name:  "list",
	Usage: "list all assigned device",
	Flags: []cli.Flag{
		cli.StringFlag{
			Name:        "sandbox-id",
			Usage:       "the target sandbox for getting the iptables",
			Required:    true,
			Destination: &sandboxID,
		},
		cli.StringFlag{
			Name:        "device-path",
			Usage:       "the absolute path of device",
			Destination: &devPath,
		},
	},
	Action: func(c *cli.Context) error {
		// verify sandbox exists:
		if err := katautils.VerifyContainerID(sandboxID); err != nil {
			return err
		}

		url := containerdshim.DeviceUrl

		body, err := shimclient.DoGet(sandboxID, defaultTimeout, url)
		if err != nil {
			return err
		}

		fmt.Println(string(body))
		return nil
	},
}

var attachDeviceCommand = cli.Command{
	Name:  "attach",
	Usage: "attach the device to sandbox",
	Flags: []cli.Flag{
		cli.StringFlag{
			Name:        "sandbox-id",
			Usage:       "the target sandbox for getting the iptables",
			Required:    true,
			Destination: &sandboxID,
		},
		cli.StringFlag{
			Name:        "device-path",
			Usage:       "the absolute path of device",
			Destination: &devPath,
		},
	},
	Action: func(c *cli.Context) error {
		// verify sandbox exists:
		if err := katautils.VerifyContainerID(sandboxID); err != nil {
			return err
		}

		resizeReq := containerdshim.DeviceRequest{
			DevicePath: devPath,
		}
		encoded, err := json.Marshal(resizeReq)
		if err != nil {
			return err
		}

		return shimclient.DoPut(sandboxID, defaultTimeout*10, containerdshim.DeviceUrl, "application/json", encoded)
	},
}

var detachDeviceCOmmand = cli.Command{
	Name:  "detach",
	Usage: "detach the device from sandbox",
	Flags: []cli.Flag{
		cli.StringFlag{
			Name:        "sandbox-id",
			Usage:       "the target sandbox for getting the iptables",
			Required:    true,
			Destination: &sandboxID,
		},
		cli.StringFlag{
			Name:        "device-path",
			Usage:       "the absolute path of device",
			Destination: &devPath,
		},
	},
	Action: func(c *cli.Context) error {
		// verify sandbox exists:
		if err := katautils.VerifyContainerID(sandboxID); err != nil {
			return err
		}

		resizeReq := containerdshim.DeviceRequest{
			DevicePath: devPath,
		}
		encoded, err := json.Marshal(resizeReq)
		if err != nil {
			return err
		}

		return shimclient.DoDelete(sandboxID, defaultTimeout*10, containerdshim.DeviceUrl, "application/json", encoded)
	},
}
