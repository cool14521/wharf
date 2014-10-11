package docker

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"net/http/httputil"
	"os"
	"strings"
	"syscall"
	"time"
	"unsafe"

	"github.com/dockercn/docker-bucket/drone/pkg/build/docker/stdcopy"
)

const (
	APIVERSION        = 1.9
	DEFAULTHTTPPORT   = 4243
	DEFAULTUNIXSOCKET = "/var/run/docker.sock"
	DEFAULTPROTOCOL   = "unix"
	DEFAULTTAG        = "latest"
	VERSION           = "0.8.0"
)

// Enables verbose logging to the Terminal window
var Logging = true

// New creates an instance of the Docker Client
func New() *Client {
	c := &Client{}

	c.setHost(DEFAULTUNIXSOCKET)

	c.Images = &ImageService{c}
	c.Containers = &ContainerService{c}
	return c
}

type Client struct {
	proto string
	addr  string

	Images     *ImageService
	Containers *ContainerService
}

var (
	// Returned if the specified resource does not exist.
	ErrNotFound = errors.New("Not Found")

	// Returned if the caller attempts to make a call or modify a resource
	// for which the caller is not authorized.
	//
	// The request was a valid request, the caller's authentication credentials
	// succeeded but those credentials do not grant the caller permission to
	// access the resource.
	ErrForbidden = errors.New("Forbidden")

	// Returned if the call requires authentication and either the credentials
	// provided failed or no credentials were provided.
	ErrNotAuthorized = errors.New("Unauthorized")

	// Returned if the caller submits a badly formed request. For example,
	// the caller can receive this return if you forget a required parameter.
	ErrBadRequest = errors.New("Bad Request")
)

func (c *Client) setHost(defaultUnixSocket string) {
	c.proto = DEFAULTPROTOCOL
	c.addr = defaultUnixSocket

	if os.Getenv("DOCKER_HOST") != "" {
		pieces := strings.Split(os.Getenv("DOCKER_HOST"), "://")
		if len(pieces) == 2 {
			c.proto = pieces[0]
			c.addr = pieces[1]
		} else if len(pieces) == 1 {
			c.addr = pieces[0]
		}
	} else {
		// if the default socket doesn't exist then
		// we'll try to connect to the default tcp address
		if _, err := os.Stat(defaultUnixSocket); err != nil {
			c.proto = "tcp"
			c.addr = "0.0.0.0:4243"
		}
	}
}

// helper function used to make HTTP requests to the Docker daemon.
func (c *Client) do(method, path string, in, out interface{}) error {
	// if data input is provided, serialize to JSON
	var payload io.Reader
	if in != nil {
		buf, err := json.Marshal(in)
		if err != nil {
			return err
		}
		payload = bytes.NewBuffer(buf)
	}

	// create the request
	req, err := http.NewRequest(method, fmt.Sprintf("/v%g%s", APIVERSION, path), payload)
	if err != nil {
		return err
	}

	// set the appropariate headers
	req.Header = http.Header{}
	req.Header.Set("User-Agent", "Docker-Client/"+VERSION)
	req.Header.Set("Content-Type", "application/json")

	// dial the host server
	req.Host = c.addr
	dial, err := net.Dial(c.proto, c.addr)
	if err != nil {
		return err
	}

	// make the request
	conn := httputil.NewClientConn(dial, nil)
	resp, err := conn.Do(req)
	defer conn.Close()
	if err != nil {
		return err
	}

	// Read the bytes from the body (make sure we defer close the body)
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	// Check for an http error status (ie not 200 StatusOK)
	switch resp.StatusCode {
	case 404:
		return ErrNotFound
	case 403:
		return ErrForbidden
	case 401:
		return ErrNotAuthorized
	case 400:
		return ErrBadRequest
	}

	// Unmarshall the JSON response
	if out != nil {
		return json.Unmarshal(body, out)
	}

	return nil
}

func (c *Client) hijack(method, path string, setRawTerminal bool, out io.Writer) error {
	req, err := http.NewRequest(method, fmt.Sprintf("/v%g%s", APIVERSION, path), nil)
	if err != nil {
		return err
	}

	req.Header.Set("User-Agent", "Docker-Client/"+VERSION)
	req.Header.Set("Content-Type", "plain/text")
	req.Host = c.addr

	dial, err := net.Dial(c.proto, c.addr)
	if err != nil {
		if strings.Contains(err.Error(), "connection refused") {
			return fmt.Errorf("Can't connect to docker daemon. Is 'docker -d' running on this host?")
		}
		return err
	}
	clientconn := httputil.NewClientConn(dial, nil)
	defer clientconn.Close()

	// Server hijacks the connection, error 'connection closed' expected
	clientconn.Do(req)

	// Hijack the connection to read / write
	rwc, br := clientconn.Hijack()
	defer rwc.Close()

	// launch a goroutine to copy the stream
	// of build output to the writer.
	errStdout := make(chan error, 1)
	go func() {
		var err error
		if setRawTerminal {
			_, err = io.Copy(out, br)
		} else {
			//_, err = utils.StdCopy(out, out, br)
			_, err = stdcopy.StdCopy(out, out, br)
		}

		errStdout <- err
	}()

	// wait for a response
	if err := <-errStdout; err != nil {
		return err
	}
	return nil
}

func (c *Client) stream(method, path string, in io.Reader, out io.Writer, headers http.Header) error {
	if (method == "POST" || method == "PUT") && in == nil {
		in = bytes.NewReader(nil)
	}

	// setup the request
	req, err := http.NewRequest(method, fmt.Sprintf("/v%g%s", APIVERSION, path), in)
	if err != nil {
		return err
	}

	// set default headers
	req.Header = headers
	req.Header.Set("User-Agent", "Docker-Client/0.6.4")
	req.Header.Set("Content-Type", "plain/text")

	// dial the host server
	req.Host = c.addr
	dial, err := net.Dial(c.proto, c.addr)
	if err != nil {
		return err
	}

	// make the request
	conn := httputil.NewClientConn(dial, nil)
	resp, err := conn.Do(req)
	defer conn.Close()
	if err != nil {
		return err
	}

	// make sure we defer close the body
	defer resp.Body.Close()

	// Check for an http error status (ie not 200 StatusOK)
	switch resp.StatusCode {
	case 404:
		return ErrNotFound
	case 403:
		return ErrForbidden
	case 401:
		return ErrNotAuthorized
	case 400:
		return ErrBadRequest
	}

	// If no output we exit now with no errors
	if out == nil {
		return nil
	}

	// copy the output stream to the writer
	if resp.Header.Get("Content-Type") == "application/json" {
		var terminalFd = os.Stdin.Fd()
		var isTerminal = IsTerminal(terminalFd)

		// it may not make sense to put this code here, but it works for
		// us at the moment, and I don't feel like refactoring
		return DisplayJSONMessagesStream(resp.Body, out, terminalFd, isTerminal)
	}
	// otherwise plain text
	if _, err := io.Copy(out, resp.Body); err != nil {
		return err
	}

	return nil
}

type JSONError struct {
	Code    int    `json:"code,omitempty"`
	Message string `json:"message,omitempty"`
}

func (e *JSONError) Error() string {
	return e.Message
}

type JSONProgress struct {
	terminalFd uintptr
	Current    int   `json:"current,omitempty"`
	Total      int   `json:"total,omitempty"`
	Start      int64 `json:"start,omitempty"`
}
type JSONMessage struct {
	Stream          string        `json:"stream,omitempty"`
	Status          string        `json:"status,omitempty"`
	Progress        *JSONProgress `json:"progressDetail,omitempty"`
	ProgressMessage string        `json:"progress,omitempty"` //deprecated
	ID              string        `json:"id,omitempty"`
	From            string        `json:"from,omitempty"`
	Time            int64         `json:"time,omitempty"`
	Error           *JSONError    `json:"errorDetail,omitempty"`
	ErrorMessage    string        `json:"error,omitempty"` //deprecated
}

func DisplayJSONMessagesStream(in io.Reader, out io.Writer, terminalFd uintptr, isTerminal bool) error {
	var (
		dec  = json.NewDecoder(in)
		ids  = make(map[string]int)
		diff = 0
	)
	for {
		var jm JSONMessage
		if err := dec.Decode(&jm); err != nil {
			if err == io.EOF {
				break
			}
			return err
		}

		if jm.Progress != nil {
			jm.Progress.terminalFd = terminalFd
		}
		if jm.ID != "" && (jm.Progress != nil || jm.ProgressMessage != "") {
			line, ok := ids[jm.ID]
			if !ok {
				line = len(ids)
				ids[jm.ID] = line
				if isTerminal {
					fmt.Fprintf(out, "\n")
				}
				diff = 0
			} else {
				diff = len(ids) - line
			}
			if jm.ID != "" && isTerminal {
				// <ESC>[{diff}A = move cursor up diff rows
				fmt.Fprintf(out, "%c[%dA", 27, diff)
			}
		}
		err := jm.Display(out, isTerminal)
		if jm.ID != "" && isTerminal {
			// <ESC>[{diff}B = move cursor down diff rows
			fmt.Fprintf(out, "%c[%dB", 27, diff)
		}
		if err != nil {
			return err
		}
	}
	return nil
}

func (jm *JSONMessage) Display(out io.Writer, isTerminal bool) error {
	if jm.Error != nil {
		if jm.Error.Code == 401 {
			return fmt.Errorf("Authentication is required.")
		}
		return jm.Error
	}
	var endl string
	if isTerminal && jm.Stream == "" && jm.Progress != nil {
		// <ESC>[2K = erase entire current line
		fmt.Fprintf(out, "%c[2K\r", 27)
		endl = "\r"
	} else if jm.Progress != nil { //disable progressbar in non-terminal
		return nil
	}
	if jm.Time != 0 {
		fmt.Fprintf(out, "%s ", time.Unix(jm.Time, 0).Format("2006-01-02T15:04:05.000000000Z07:00"))
	}
	if jm.ID != "" {
		fmt.Fprintf(out, "%s: ", jm.ID)
	}
	if jm.From != "" {
		fmt.Fprintf(out, "(from %s) ", jm.From)
	}
	if jm.Progress != nil {
		fmt.Fprintf(out, "%s %s%s", jm.Status, jm.Progress.String(), endl)
	} else if jm.ProgressMessage != "" { //deprecated
		fmt.Fprintf(out, "%s %s%s", jm.Status, jm.ProgressMessage, endl)
	} else if jm.Stream != "" {
		fmt.Fprintf(out, "%s%s", jm.Stream, endl)
	} else {
		fmt.Fprintf(out, "%s%s\n", jm.Status, endl)
	}
	return nil
}

func (p *JSONProgress) String() string {
	var (
		width       = 200
		pbBox       string
		numbersBox  string
		timeLeftBox string
	)

	ws, err := GetWinsize(p.terminalFd)
	if err == nil {
		width = int(ws.Width)
	}

	if p.Current <= 0 && p.Total <= 0 {
		return ""
	}
	current := HumanSize(int64(p.Current))
	if p.Total <= 0 {
		return fmt.Sprintf("%8v", current)
	}
	total := HumanSize(int64(p.Total))
	percentage := int(float64(p.Current)/float64(p.Total)*100) / 2
	if width > 110 {
		// this number can't be negetive gh#7136
		numSpaces := 0
		if 50-percentage > 0 {
			numSpaces = 50 - percentage
		}
		pbBox = fmt.Sprintf("[%s>%s] ", strings.Repeat("=", percentage), strings.Repeat(" ", numSpaces))
	}
	numbersBox = fmt.Sprintf("%8v/%v", current, total)

	if p.Current > 0 && p.Start > 0 && percentage < 50 {
		fromStart := time.Now().UTC().Sub(time.Unix(int64(p.Start), 0))
		perEntry := fromStart / time.Duration(p.Current)
		left := time.Duration(p.Total-p.Current) * perEntry
		left = (left / time.Second) * time.Second

		if width > 50 {
			timeLeftBox = " " + left.String()
		}
	}
	return pbBox + numbersBox + timeLeftBox
}

var unitAbbrs = [...]string{"B", "kB", "MB", "GB", "TB", "PB", "EB", "ZB", "YB"}

// HumanSize returns a human-readable approximation of a size
// using SI standard (eg. "44kB", "17MB")
func HumanSize(size int64) string {
	i := 0
	sizef := float64(size)
	for sizef >= 1000.0 {
		sizef = sizef / 1000.0
		i++
	}
	return fmt.Sprintf("%.4g %s", sizef, unitAbbrs[i])
}

type Termios struct {
	Iflag  uint32
	Oflag  uint32
	Cflag  uint32
	Lflag  uint32
	Cc     [20]byte
	Ispeed uint32
	Ospeed uint32
}

const (
	getTermios = syscall.TCGETS
)

// IsTerminal returns true if the given file descriptor is a terminal.
func IsTerminal(fd uintptr) bool {
	var termios Termios
	_, _, err := syscall.Syscall(syscall.SYS_IOCTL, fd, uintptr(getTermios), uintptr(unsafe.Pointer(&termios)))
	return err == 0
}

type Winsize struct {
	Height uint16
	Width  uint16
	x      uint16
	y      uint16
}

func GetWinsize(fd uintptr) (*Winsize, error) {
	ws := &Winsize{}
	_, _, err := syscall.Syscall(syscall.SYS_IOCTL, fd, uintptr(syscall.TIOCGWINSZ), uintptr(unsafe.Pointer(ws)))
	// Skipp errno = 0
	if err == 0 {
		return ws, nil
	}
	return ws, err
}
