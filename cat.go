package main

import (
	"io"

	"github.com/anchor/dataframe"
)

func frameCat(cfg *Config, r FrameReader, w Writer) int {
	var encoder FrameEncoder
	switch cfg.Output.Format {
	case "raw":
		e := new(RawFrameEncoder)
		encoder = *e
	case "json":
		e := new(JsonFrameEncoder)
		encoder = *e
	default:
		Errorf("Invalid output format %s.", cfg.Output.Format)
	}
	var f *dataframe.DataFrame
	var err error
	for f, err = r.NextFrame(); err == nil ; f, err = r.NextFrame() {
		b, err := encoder.EncodeFrame(f)
		if err != nil {
			Errorf("Error encoding frame: %v", err)
			return 1
		}
		err = w.Write(b)
		if err != nil {
			Errorf("Error writing frame: %v", err)
		}
	}
	if err != io.EOF {
		Errorf("Error reading next frame: %v", err)
	}
	return 0
}

func burstCat(cfg *Config, r FrameReader, w Writer) int {
	var encoder BurstEncoder
	switch cfg.Output.Format {
	case "raw":
		e := new(RawBurstEncoder)
		encoder = *e
	}
	fs := make([]*dataframe.DataFrame, 0)
	var f *dataframe.DataFrame
	var err error
	for f, err = r.NextFrame(); err == nil ; f, err = r.NextFrame() {
		fs = append(fs, f)
	}
	if err != io.EOF {
		Errorf("Error reading next frame: %v", err)
	}
	b, err := encoder.EncodeBurst(dataframe.BuildDataBurst(fs))
	if err != nil {
		Errorf("Error marshalling databurst: %v", err)
	}
	w.Write(b)
	return 0
}

func CatCommand(cfg *Config, r FrameReader, w Writer) int {
	if cfg.Output.Packing == BurstPacking {
		return burstCat(cfg, r, w)
	} else {
		return frameCat(cfg, r, w)
	}
	return 0
}
