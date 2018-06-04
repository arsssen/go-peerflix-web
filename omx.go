package main

import (
	"fmt"
	"io"
	"log"
	"os/exec"
)

var commandList = map[string]string{
	"playpause": "p",
	"subn":      "m",
	"quit":      "q",
	"info":      "z",
	"subs":      "s",
	"backward":  "\x1b[D",
	"forward":   "\x1b[C",
}

//Omx is the omxplayer control struct
type Omx struct {
	Command *exec.Cmd
	PipeIn  io.WriteCloser
	PipeOut io.ReadCloser
	Playing bool
}

var omxPlayer Omx

// Start will Start omxplayer playback for a given filename
// it will stop any playback that is currently running
func (p *Omx) Start(filename string) error {
	var err error
	if p.Playing == true {
		p.SendCommand("stop")
		p.Playing = false
	}
	log.Printf("-omx playing %s", filename)
	p.Command = exec.Command("omxplayer", "-o", "hdmi", filename)
	p.PipeIn, err = p.Command.StdinPipe()

	if err != nil {
		return err
	}

	p.PipeOut, err = p.Command.StderrPipe()

	if err != nil {
		return err
	}

	p.Playing = true
	err = p.Command.Start()

	if err != nil {
		p.Playing = false
	}

	return err
}

// SendCommand sends command to omxplayer using the pipe from the struct.
func (p *Omx) SendCommand(command string) error {
	if cmd, exists := commandList[command]; exists {
		_, err := p.PipeIn.Write([]byte(cmd))
		if err != nil {
			log.Printf("error sending cmd: %s", err)
			p.Playing = false
		}
		if cmd == commandList["quit"] {
			p.Playing = false
		}
		return err
	}
	return fmt.Errorf("unknown command: %s", command)
}
