package types

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/ligjn/aog/internal/utils"
)

type ServiceResultType int

const (
	ServiceResultDone ServiceResultType = iota
	ServiceResultFailed
	ServiceResultChunk
)

type ServiceResult struct {
	Type       ServiceResultType
	TaskId     uint64
	Error      error
	StatusCode int
	HTTP       HTTPContent
}

func IsDropAction(err error) bool {
	if err == nil {
		return false
	}
	var dropAction *DropAction
	ok := errors.As(err, &dropAction)
	return ok
}

func (sr *ServiceResult) WriteBack(w http.ResponseWriter) {
	if IsDropAction(sr.Error) {
		return
	}
	if sr.Type == ServiceResultFailed {
		if httpError, ok := sr.Error.(*HTTPErrorResponse); ok {
			clear(w.Header())
			w.WriteHeader(httpError.StatusCode)
			for k, v := range httpError.Header {
				w.Header().Set(k, v[0])
			}
			_, _ = w.Write(httpError.Body)
			// event.SysEvents.NotifyHTTPResponse("send_back_response", httpError.StatusCode, w.Header(), httpError.Body)
			return
		}

		w.WriteHeader(http.StatusInternalServerError)
		var errBytes []byte
		if sr.Error != nil {
			errBytes = []byte(sr.Error.Error())
			_, _ = w.Write(errBytes)
		}
		// event.SysEvents.NotifyHTTPResponse("send_back_response", http.StatusInternalServerError, w.Header(), errBytes)
	} else {
		clear(w.Header())
		w.WriteHeader(sr.StatusCode)
		for k, v := range sr.HTTP.Header {
			w.Header().Set(k, v[0])
		}
		// event.SysEvents.NotifyHTTPResponse("send_back_response", sr.StatusCode, w.Header(), sr.HTTP.Body)
		_, _ = w.Write(sr.HTTP.Body)
	}
}

func (sr *ServiceResult) String() string {
	var stype string
	switch sr.Type {
	case ServiceResultDone:
		stype = "ServiceResultDone"
	case ServiceResultFailed:
		stype = "ServiceResultFailed"
	case ServiceResultChunk:
		stype = "ServiceResultChunk"
	}
	return fmt.Sprintf("ServiceResult{Type: %s, TaskId: %d, StatusCode: %d, Error: %v, HTTP {Header: %+v, Body: %s}}",
		stype, sr.StatusCode, sr.TaskId, sr.Error, sr.HTTP.Header, utils.BodyToString(sr.HTTP.Header, sr.HTTP.Body))
}

// ServiceRequest The body of the OriginalRequest has been read out so need to placed here
type ServiceRequest struct {
	AskStreamMode         bool          `json:"stream"`
	Model                 string        `json:"model"`
	HybridPolicy          string        `json:"hybrid_policy"`
	RemoteServiceProvider string        `json:"remote_service_provider"`
	FromFlavor            string        `json:"-"`
	Service               string        `json:"-"`
	Priority              int           `json:"-"`
	RequestSegments       int           `json:"request_segments"`
	RequestExtraUrl       string        `json:"extra_url"`
	HTTP                  HTTPContent   `json:"-"`
	OriginalRequest       *http.Request `json:"-"`
}

func (sr *ServiceRequest) String() string {
	return fmt.Sprintf("ServiceRequest{FromFlavor: %s, Service: %s, Model: %s, Stream: %t, Hybrid: %s}",
		sr.FromFlavor, sr.Service, sr.Model, sr.AskStreamMode, sr.HybridPolicy)
}

type ServiceTarget struct {
	Location        string
	Stream          bool
	Model           string
	ToFavor         string
	XPU             string
	ServiceProvider *ServiceProvider
}

func (sr *ServiceTarget) String() string {
	return fmt.Sprintf("ServiceDispatch{Location: %s, Provider: %s}", sr.Location, sr.ServiceProvider.ProviderName)
}

type StreamModeType int

const (
	StreamModeNonStream StreamModeType = iota
	StreamModeEventStream
	StreamModeNDJson
)

func (st StreamModeType) String() string {
	switch st {
	case StreamModeNonStream:
		return "NonStream"
	case StreamModeEventStream:
		return "EventStream"
	case StreamModeNDJson:
		return "NDJson"
	}
	return "Unknown"
}

type StreamMode struct {
	Mode   StreamModeType
	Header http.Header
}

func (sm *StreamMode) IsStream() bool {
	return sm.Mode != StreamModeNonStream
}

// ReadChunk the chunks of the stream is delimited by "\n\n" for event-stream, and "\n"
// or "\r\n" for x=ndjson so only need to read to '\n'. However, reader can only
// stop at one character, so we need to handle it by ourselves for "\n\n" case
func (sm *StreamMode) ReadChunk(reader *bufio.Reader) ([]byte, error) {
	if sm.Mode == StreamModeNonStream {
		return io.ReadAll(reader)
	}

	if sm.Mode == StreamModeNDJson {
		return reader.ReadBytes('\n')
	}

	var buffer bytes.Buffer
	var line []byte
	var err error
	for {
		line, err = reader.ReadBytes('\n')
		if err != nil && err != io.EOF { // real error
			break
		}
		buffer.Write(line)
		if err == io.EOF { // the end
			break
		}
		// Check the character before '\n'
		if buffer.Len() >= 2 && buffer.Bytes()[buffer.Len()-2] == '\n' && buffer.Bytes()[buffer.Len()-1] == '\n' {
			break
		}
	}
	return buffer.Bytes(), err
}

// UnwrapChunk Get real data
// event-stream is started with "data: " which need to be removed
// TODO: event-stream may contain multiple fields, and we only need to pick
// the data: fields. Need to handle it in the future. Currently we assume there
// is only a data: field.
func (sm *StreamMode) UnwrapChunk(chunk []byte) []byte {
	// remove "data: " at the beginning of the chunk
	if sm.Mode == StreamModeEventStream {
		if len(chunk) >= 6 && bytes.HasPrefix(chunk, []byte("data: ")) {
			return chunk[6:]
		}
	}
	return chunk
}

func (sm *StreamMode) WrapChunk(chunk []byte) []byte {
	if sm.Mode == StreamModeNonStream {
		return chunk
	}
	n := 0
	for i := len(chunk) - 1; i >= 0 && chunk[i] == '\n'; i-- {
		n++
	}

	nNeed := 0
	if sm.Mode == StreamModeNDJson {
		nNeed = 1
	} else if sm.Mode == StreamModeEventStream {
		nNeed = 2
	}

	if n == nNeed {
		return chunk
	} else if n > nNeed {
		return chunk[:len(chunk)-n+nNeed]
	} else {
		for range nNeed - n {
			chunk = append(chunk, '\n')
		}
	}

	if sm.Mode == StreamModeEventStream && !bytes.HasPrefix(chunk, []byte("data: ")) {
		chunk = append([]byte("data: "), chunk...)
	}
	return chunk
}
