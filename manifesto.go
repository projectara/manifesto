/*
 * Manifesto - generates a Greybus Module Manifest from git config syntax input
 *
 * Copyright 2014 Google Inc.
 * Copyright 2014 Linaro Ltd.
 *
 * Note: quick and dirty, needs a bunch of error checking on I/O and data read
 * 	 from the gcfg-style manifest source file
 */

package main

import (
	"bufio"
	"code.google.com/p/gcfg"
	"encoding/binary"
	"path/filepath"
	"log"
	"os"
	"strings"
)

const (
	INVALID_TYPE		uint8	= iota
	MODULE_TYPE		uint8	= iota
	STRING_TYPE		uint8	= iota
	INTERFACE_TYPE		uint8	= iota
	CPORT_TYPE		uint8	= iota
	CLASS_TYPE		uint8	= iota
)

const (
	MANIFEST_HEADER_SIZE	uint16	= 0x04
	MODULE_SIZE		uint16	= 0x13
	STRING_SIZE		uint16	= 0x05
	INTERFACE_SIZE		uint16	= 0x04
	CPORT_SIZE		uint16	= 0x07
	CLASS_SIZE		uint16  = 0x04
)

// Greybus version 0.1 manifest format
type Manifest struct {
	Manifest_header struct {
		Size uint16
		Version_major uint8
		Version_minor uint8
	}
	Module_descriptor struct {
		Size uint16
		Type uint8
		Vendor uint16
		Product uint16
		Version uint16
		Vendor_string_id uint8
		Product_string_id uint8
		Unique_id uint64
	}
	Interface_descriptor map[string] *struct {
		Size uint16
		Type uint8
		Id uint8
	}
	String_descriptor map[string] *struct {
		Size uint16
		Type uint8
		Length uint8
		Id uint8
		String string
	}
	Cport_descriptor map[string] *struct {
		Size uint16
		Type uint8
		Interface uint8
		Id uint16
		Protocol uint8
	}
	Class_descriptor map[string] *struct {
		Size uint16
		Type uint8
		Class uint8
	}
}

func right_pad(s string, padStr string, pLen int) string {
	return s + strings.Repeat(padStr, pLen);
}

func populate_manifest(mnf Manifest) Manifest {
	var mnf_size uint16
	mnf_size = 0

	mnf.Module_descriptor.Type = MODULE_TYPE
	mnf.Module_descriptor.Size = MODULE_SIZE
	mnf_size = mnf_size +
		(uint16)(mnf.Module_descriptor.Size)

	for k := range mnf.String_descriptor {
		var size uint16

		/* Raw string length for the length field */
		length := len(mnf.String_descriptor[k].String)
		mnf.String_descriptor[k].Length = (uint8)(length)
		/* Pad strings to 4 byte alignment */
		mnf.String_descriptor[k].String =
			right_pad(mnf.String_descriptor[k].String,
				 "\x00", length % 4)
		/* Total string descriptor size includes string pad */
		size = STRING_SIZE +
			(uint16)(len(mnf.String_descriptor[k].String))
		mnf.String_descriptor[k].Type = STRING_TYPE
		mnf.String_descriptor[k].Size = size
		mnf_size = mnf_size + (uint16)(size)
	}

	for k := range mnf.Interface_descriptor {
		mnf.Interface_descriptor[k].Type = INTERFACE_TYPE
		mnf.Interface_descriptor[k].Size = INTERFACE_SIZE
		mnf_size = mnf_size +
			(uint16)(mnf.Interface_descriptor[k].Size)
	}

	for k := range mnf.Cport_descriptor {
		mnf.Cport_descriptor[k].Type = CPORT_TYPE
		mnf.Cport_descriptor[k].Size = CPORT_SIZE
		mnf_size = mnf_size +
			(uint16)(mnf.Cport_descriptor[k].Size)
	}

	for k := range mnf.Class_descriptor {
		mnf.Class_descriptor[k].Type = CLASS_TYPE
		mnf.Class_descriptor[k].Size = CLASS_SIZE
		mnf_size = mnf_size +
			(uint16)(mnf.Class_descriptor[k].Size)
	}

	/* Total size of all descriptors plus our header */
	mnf.Manifest_header.Size = MANIFEST_HEADER_SIZE + mnf_size

	return mnf
}

func write_manifest(m *os.File, mnf Manifest) {
	mwriter := bufio.NewWriter(m)

	/* Manifest header */
	binary.Write(mwriter, binary.LittleEndian, mnf.Manifest_header)

	/* Module descriptor */
	binary.Write(mwriter, binary.LittleEndian, mnf.Module_descriptor)

	/* Cport descriptors */
	for k := range mnf.Cport_descriptor {
		binary.Write(mwriter, binary.LittleEndian,
			     mnf.Cport_descriptor[k])
	}

	/* Interface descriptors */
	for k := range mnf.Interface_descriptor {
		binary.Write(mwriter, binary.LittleEndian,
			     mnf.Interface_descriptor[k])
	}

	/* Class descriptors */
	for k := range mnf.Class_descriptor {
		binary.Write(mwriter, binary.LittleEndian,
			     mnf.Class_descriptor[k])
	}

	/* String descriptors */
	for k := range mnf.String_descriptor {
		strdesc := mnf.String_descriptor[k]

		/* Binary writer doesn't work on strings, so do it manually */
		binary.Write(mwriter, binary.LittleEndian, strdesc.Size)
		binary.Write(mwriter, binary.LittleEndian, strdesc.Type)
		binary.Write(mwriter, binary.LittleEndian, strdesc.Length)
		binary.Write(mwriter, binary.LittleEndian, strdesc.Id)
		mwriter.WriteString(strdesc.String)
	}

	mwriter.Flush()
}


func main() {
	var mnf Manifest

	// Manifest source file is our only argument
	mnfs := os.Args[1];

	// TODO: error checking on args

	// Open a binary manifest output file
	basename := mnfs[0:len(mnfs)-len(filepath.Ext(mnfs))]
	mnfb := basename + ".mnfb"
	m, err := os.Create(mnfb)
	if err != nil {
		panic(err)
	}
	defer m.Close()

	// Read in static manifest fields
	err = gcfg.ReadFileInto(&mnf, mnfs)
	if err != nil {
		log.Fatalf("Failed to parse manifest source: %s", err)
	}

	// Populate calculated manifest fields
	mnf = populate_manifest(mnf)

	// Write out manifest
	write_manifest(m, mnf);
}

