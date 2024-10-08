/*
Copyright 2021 IBM All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package gateway

import (
	"fmt"
	"strings"

	gp "github.com/hyperledger/fabric-protos-go-apiv2/gateway"
	"github.com/hyperledger/fabric-protos-go-apiv2/peer"
	"github.com/hyperledger/fabric/core/chaincode"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/protoadapt"
)

// responseStatus unpacks the proposal response and error values that are returned from ProcessProposal and
// determines how the gateway should react (retry?, close connection?).
// Uses the grpc canonical status error codes and their recommended actions.
// Returns:
// - response status code, with codes.OK indicating success and other values indicating likely error type
// - error message extracted from the err or generated from 500 proposal response (string)
// - should the gateway retry (only the Evaluate() uses this) (bool)
// - should the gateway close the connection and remove the peer from its registry (bool)
func responseStatus(response *peer.ProposalResponse, err error) (statusCode codes.Code, message string, retry bool, remove bool) {
	if err != nil {
		if response == nil {
			// there is no ProposalResponse, so this must have been generated by grpc in response to an unavailable peer
			// - close the connection and retry on another
			return codes.Unavailable, err.Error(), true, true
		}
		// there is a response and an err, so it must have been from the unpackProposal() or preProcess() stages
		// preProcess does all the signature and ACL checking. In either case, no point retrying, or closing the connection (it's a client error)
		return codes.FailedPrecondition, err.Error(), false, false
	}
	if response.Response.Status < 200 || response.Response.Status >= 400 {
		if response.Payload == nil && response.Response.Status == 500 {
			// there's a error 500 response but no payload, so the response was generated in the peer rather than the chaincode
			if strings.HasSuffix(response.Response.Message, chaincode.ErrorStreamTerminated) {
				// chaincode container crashed probably. Close connection and retry on another peer
				return codes.Aborted, response.Response.Message, true, true
			}
			// some other error - retry on another peer
			return codes.Aborted, response.Response.Message, true, false
		} else {
			// otherwise it must be an error response generated by the chaincode
			return codes.Unknown, fmt.Sprintf("chaincode response %d, %s", response.Response.Status, response.Response.Message), false, false
		}
	}
	// anything else is a success
	return codes.OK, "", false, false
}

func newRpcError(code codes.Code, message string, details ...proto.Message) error {
	st := status.New(code, message)
	if len(details) != 0 {
		var ds []protoadapt.MessageV1
		for _, detail := range details {
			d := protoadapt.MessageV1Of(detail)
			ds = append(ds, d)
		}

		std, err := st.WithDetails(ds...)
		if err == nil {
			return std.Err()
		} // otherwise return the error without the details
	}
	return st.Err()
}

func toRpcError(err error, unknownCode codes.Code) error {
	errStatus := toRpcStatus(err)
	if errStatus.Code() != codes.Unknown {
		return errStatus.Err()
	}

	return status.Error(unknownCode, err.Error())
}

func toRpcStatus(err error) *status.Status {
	errStatus, ok := status.FromError(err)
	if ok {
		return errStatus
	}

	return status.FromContextError(err)
}

func errorDetail(e *endpointConfig, msg string) *gp.ErrorDetail {
	return &gp.ErrorDetail{Address: e.logAddress, MspId: e.mspid, Message: msg}
}
