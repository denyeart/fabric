/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package dispatcher_test

import (
	"fmt"
	"time"

	"github.com/hyperledger/fabric/core/dispatcher"
	"github.com/hyperledger/fabric/core/dispatcher/mock"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type TestReceiver struct{}

func (tr TestReceiver) GoodFunc(ts *timestamppb.Timestamp) (*timestamppb.Timestamp, error) {
	return timestamppb.New(time.Unix(0, 0)), nil
}

func (tr TestReceiver) MissingFuncParameters() (*timestamppb.Timestamp, error) {
	return timestamppb.New(time.Unix(0, 0)), nil
}

func (tr TestReceiver) NotProtoParameter(foo *string) (*timestamppb.Timestamp, error) {
	return timestamppb.New(time.Unix(0, 0)), nil
}

func (tr TestReceiver) NotPointerParameter(foo string) (*timestamppb.Timestamp, error) {
	return timestamppb.New(time.Unix(0, 0)), nil
}

func (tr TestReceiver) NoReturnValues(ts *timestamppb.Timestamp) {}

func (tr TestReceiver) NotProtoReturn(ts *timestamppb.Timestamp) (string, error) {
	return "", nil
}

func (tr TestReceiver) NotErrorReturn(ts *timestamppb.Timestamp) (*timestamppb.Timestamp, string) {
	return nil, ""
}

func (tr TestReceiver) NilNilReturn(ts *timestamppb.Timestamp) (*timestamppb.Timestamp, error) {
	return nil, nil
}

func (tr TestReceiver) ErrorReturned(ts *timestamppb.Timestamp) (*timestamppb.Timestamp, error) {
	return nil, fmt.Errorf("fake-error")
}

var _ = Describe("Dispatcher", func() {
	var (
		d         *dispatcher.Dispatcher
		fakeProto *mock.Protobuf
	)

	BeforeEach(func() {
		fakeProto = &mock.Protobuf{}
		fakeProto.MarshalStub = proto.Marshal
		fakeProto.UnmarshalStub = proto.Unmarshal

		d = &dispatcher.Dispatcher{
			Protobuf: fakeProto,
		}
	})

	Describe("Dispatch", func() {
		var (
			testReceiver TestReceiver
			inputBytes   []byte
		)

		BeforeEach(func() {
			var err error
			inputBytes, err = proto.Marshal(timestamppb.Now())
			Expect(err).NotTo(HaveOccurred())
		})

		It("unmarshals, dispatches to the correct function, and marshals the result", func() {
			outputBytes, err := d.Dispatch(inputBytes, "GoodFunc", testReceiver)
			Expect(err).NotTo(HaveOccurred())
			ts := &timestamppb.Timestamp{}
			err = proto.Unmarshal(outputBytes, ts)
			Expect(err).NotTo(HaveOccurred())
			gts := ts.AsTime()
			Expect(gts).To(Equal(time.Unix(0, 0).UTC()))
		})

		Context("when the receiver does not have a method to dispatch to", func() {
			It("returns an error", func() {
				_, err := d.Dispatch(inputBytes, "MissingMethod", testReceiver)
				Expect(err).To(MatchError("receiver dispatcher_test.TestReceiver.MissingMethod does not exist"))
			})
		})

		Context("when the receiver does not return the right number of parameters", func() {
			It("returns an error", func() {
				_, err := d.Dispatch(inputBytes, "MissingFuncParameters", testReceiver)
				Expect(err).To(MatchError("receiver dispatcher_test.TestReceiver.MissingFuncParameters has 0 parameters but expected 1"))
			})
		})

		Context("when the receiver does not take a pointer", func() {
			It("returns an error", func() {
				_, err := d.Dispatch(inputBytes, "NotPointerParameter", testReceiver)
				Expect(err).To(MatchError("receiver dispatcher_test.TestReceiver.NotPointerParameter does not accept a pointer as its argument"))
			})
		})

		Context("when the receiver does not take a protobuf message", func() {
			It("returns an error", func() {
				_, err := d.Dispatch(inputBytes, "NotProtoParameter", testReceiver)
				Expect(err).To(MatchError("receiver dispatcher_test.TestReceiver.NotProtoParameter does not accept a proto.Message as its argument, it is '*string'"))
			})
		})

		Context("when the input bytes cannot be unmarshaled", func() {
			It("wraps and returns the error", func() {
				_, err := d.Dispatch([]byte("garbage"), "GoodFunc", testReceiver)
				Expect(err).To(Not(BeNil()))
				Expect(err.Error()).To(HavePrefix("could not decode input arg for dispatcher_test.TestReceiver.GoodFunc"))
			})
		})

		Context("when the receiver does not return the right number of parameters", func() {
			It("returns an error", func() {
				_, err := d.Dispatch(inputBytes, "NoReturnValues", testReceiver)
				Expect(err).To(MatchError("receiver dispatcher_test.TestReceiver.NoReturnValues returns 0 values but expected 2"))
			})
		})

		Context("when the receiver does not return a proto message as the first return value", func() {
			It("returns an error", func() {
				_, err := d.Dispatch(inputBytes, "NotProtoReturn", testReceiver)
				Expect(err).To(MatchError("receiver dispatcher_test.TestReceiver.NotProtoReturn does not return a an implementor of proto.Message as its first return value"))
			})
		})

		Context("when the receiver does not return an error as its second return value", func() {
			It("returns an error", func() {
				_, err := d.Dispatch(inputBytes, "NotErrorReturn", testReceiver)
				Expect(err).To(MatchError("receiver dispatcher_test.TestReceiver.NotErrorReturn does not return an error as its second return value"))
			})
		})

		Context("when the receiver returns nil, nil", func() {
			It("returns an error", func() {
				_, err := d.Dispatch(inputBytes, "NilNilReturn", testReceiver)
				Expect(err).To(MatchError("receiver dispatcher_test.TestReceiver.NilNilReturn returned (nil, nil) which is not allowed"))
			})
		})

		Context("when the receiver returns an error", func() {
			It("returns the error", func() {
				_, err := d.Dispatch(inputBytes, "ErrorReturned", testReceiver)
				Expect(err).To(MatchError("fake-error"))
			})
		})

		Context("when the returned output cannot be marshaled", func() {
			BeforeEach(func() {
				fakeProto.MarshalReturns(nil, fmt.Errorf("fake-error"))
			})

			It("wraps and returns the error", func() {
				_, err := d.Dispatch(inputBytes, "GoodFunc", testReceiver)
				Expect(err).To(MatchError("failed to marshal result for dispatcher_test.TestReceiver.GoodFunc: fake-error"))
			})
		})
	})
})
