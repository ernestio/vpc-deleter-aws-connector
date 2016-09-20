/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/. */

package main

import (
	"fmt"
	"os"
	"runtime"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/nats-io/nats"

	ecc "github.com/ernestio/ernest-config-client"
)

var nc *nats.Conn
var natsErr error

func eventHandler(m *nats.Msg) {
	var n Event

	err := n.Process(m.Data)
	if err != nil {
		return
	}

	if err = n.Validate(); err != nil {
		n.Error(err)
		return
	}

	err = deleteVPC(&n)
	if err != nil {
		n.Error(err)
		return
	}

	n.Complete()
}

func deleteVPC(ev *Event) error {
	creds := credentials.NewStaticCredentials(ev.DatacenterAccessKey, ev.DatacenterAccessToken, "")
	svc := ec2.New(session.New(), &aws.Config{
		Region:      aws.String(ev.DatacenterRegion),
		Credentials: creds,
	})

	req := ec2.DeleteVpcInput{
		VpcId: aws.String(ev.VpcID),
	}
	_, err := svc.DeleteVpc(&req)
	if err != nil {
		ev.ErrorMessage = "WARN : Could not remove the vpc - " + err.Error()
		return nil
	}

	return nil
}

func main() {
	nc = ecc.NewConfig(os.Getenv("NATS_URI")).Nats()

	fmt.Println("listening for vpc.delete.aws")
	nc.Subscribe("vpc.delete.aws", eventHandler)

	runtime.Goexit()
}
