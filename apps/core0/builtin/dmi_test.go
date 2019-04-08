package builtin

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

const (
	sample1 = `
# dmidecode 3.1
Getting SMBIOS data from sysfs.
SMBIOS 2.6 present.

Handle 0x0001, DMI type 1, 27 bytes
System Information
		Manufacturer: LENOVO
		Product Name: 20042
		Version: Lenovo G560
		Serial Number: 2677240001087
		UUID: CB3E6A50-A77B-E011-88E9-B870F4165734
		Wake-up Type: Power Switch
		SKU Number: Calpella_CRB
		Family: Intel_Mobile
	`
	sample2 = `
Getting SMBIOS data from sysfs.
SMBIOS 2.6 present.

Handle 0x0000, DMI type 0, 24 bytes
BIOS Information
		Vendor: LENOVO
		Version: 29CN40WW(V2.17)
		Release Date: 04/13/2011
		ROM Size: 2048 kB
		Characteristics:
				PCI is supported
				BIOS is upgradeable
				BIOS shadowing is allowed
				Boot from CD is supported
				Selectable boot is supported
				EDD is supported
				Japanese floppy for NEC 9800 1.2 MB is supported (int 13h)
				Japanese floppy for Toshiba 1.2 MB is supported (int 13h)
				5.25"/360 kB floppy services are supported (int 13h)
				5.25"/1.2 MB floppy services are supported (int 13h)
				3.5"/720 kB floppy services are supported (int 13h)
				3.5"/2.88 MB floppy services are supported (int 13h)
				8042 keyboard services are supported (int 9h)
				CGA/mono video services are supported (int 10h)
				ACPI is supported
				USB legacy is supported
				BIOS boot specification is supported
				Targeted content distribution is supported
		BIOS Revision: 1.40
	`
	sample3 = `
# dmidecode 3.1
Getting SMBIOS data from sysfs.
SMBIOS 2.6 present.

Handle 0x0001, DMI type 1, 27 bytes
System Information
		Manufacturer: LENOVO
		Product Name: 20042
		Version: Lenovo G560
		Serial Number: 2677240001087
		UUID: CB3E6A50-A77B-E011-88E9-B870F4165734
		Wake-up Type: Power Switch
		SKU Number: Calpella_CRB
		Family: Intel_Mobile

Handle 0x000D, DMI type 12, 5 bytes
System Configuration Options
		Option 1: String1 for Type12 Equipment Manufacturer
		Option 2: String2 for Type12 Equipment Manufacturer
		Option 3: String3 for Type12 Equipment Manufacturer
		Option 4: String4 for Type12 Equipment Manufacturer

Handle 0x000E, DMI type 15, 29 bytes
System Event Log
		Area Length: 0 bytes
		Header Start Offset: 0x0000
		Data Start Offset: 0x0000
		Access Method: General-purpose non-volatile data functions
		Access Address: 0x0000
		Status: Valid, Not Full
		Change Token: 0x12345678
		Header Format: OEM-specific
		Supported Log Type Descriptors: 3
		Descriptor 1: POST memory resize
		Data Format 1: None
		Descriptor 2: POST error
		Data Format 2: POST results bitmap
		Descriptor 3: Log area reset/cleared
		Data Format 3: None

Handle 0x0011, DMI type 32, 20 bytes
System Boot Information
		Status: No errors detected
	`
	sample4 = `
# dmidecode 3.1
Getting SMBIOS data from sysfs.
SMBIOS 2.6 present.

Handle 0x0000, DMI type 0, 24 bytes
BIOS Information
		Vendor: LENOVO
		Version: 29CN40WW(V2.17)
		Release Date: 04/13/2011
		ROM Size: 2048 kB
		Characteristics:
				PCI is supported
				BIOS is upgradeable
				BIOS shadowing is allowed
				Boot from CD is supported
				Selectable boot is supported
				EDD is supported
				Japanese floppy for NEC 9800 1.2 MB is supported (int 13h)
				Japanese floppy for Toshiba 1.2 MB is supported (int 13h)
				5.25"/360 kB floppy services are supported (int 13h)
				5.25"/1.2 MB floppy services are supported (int 13h)
				3.5"/720 kB floppy services are supported (int 13h)
				3.5"/2.88 MB floppy services are supported (int 13h)
				8042 keyboard services are supported (int 9h)
				CGA/mono video services are supported (int 10h)
				ACPI is supported
				USB legacy is supported
				BIOS boot specification is supported
				Targeted content distribution is supported
		BIOS Revision: 1.40

Handle 0x002C, DMI type 4, 42 bytes
Processor Information
		Socket Designation: CPU
		Type: Central Processor
		Family: Core 2 Duo
		Manufacturer: Intel(R) Corporation
		ID: 55 06 02 00 FF FB EB BF
		Signature: Type 0, Family 6, Model 37, Stepping 5
		Flags:
				FPU (Floating-point unit on-chip)
				VME (Virtual mode extension)
				DE (Debugging extension)
				PSE (Page size extension)
				TSC (Time stamp counter)
				MSR (Model specific registers)
				PAE (Physical address extension)
				MCE (Machine check exception)
				CX8 (CMPXCHG8 instruction supported)
				APIC (On-chip APIC hardware supported)
				SEP (Fast system call)
				MTRR (Memory type range registers)
				PGE (Page global enable)
				MCA (Machine check architecture)
				CMOV (Conditional move instruction supported)
				PAT (Page attribute table)
				PSE-36 (36-bit page size extension)
				CLFSH (CLFLUSH instruction supported)
				DS (Debug store)
				ACPI (ACPI supported)
				MMX (MMX technology supported)
				FXSR (FXSAVE and FXSTOR instructions supported)
				SSE (Streaming SIMD extensions)
				SSE2 (Streaming SIMD extensions 2)
				SS (Self-snoop)
				HTT (Multi-threading)
				TM (Thermal monitor supported)
				PBE (Pending break enabled)
		Version: Intel(R) Core(TM) i3 CPU       M 370  @ 2.40GHz
		Voltage: 0.0 V
		External Clock: 1066 MHz
		Max Speed: 2400 MHz
		Current Speed: 2399 MHz
		Status: Populated, Enabled
		Upgrade: ZIF Socket
		L1 Cache Handle: 0x0030
		L2 Cache Handle: 0x002F
		L3 Cache Handle: 0x002D
		Serial Number: Not Specified
		Asset Tag: FFFF
		Part Number: Not Specified
		Core Count: 2
		Core Enabled: 2
		Thread Count: 4
		Characteristics:
				64-bit capable
	`

	sample5 = `
# dmidecode 3.1
Getting SMBIOS data from sysfs.
SMBIOS 3.0.0 present.
Table at 0x7AEAA000.

Handle 0x0014, DMI type 10, 20 bytes
On Board Device 1 Information
	Type: Video
	Status: Enabled
	Description:  Intel(R) HD Graphics Device
On Board Device 2 Information
	Type: Ethernet
	Status: Enabled
	Description:  Intel(R) I219-V Gigabit Network Device
On Board Device 3 Information
	Type: Sound
	Status: Enabled
	Description:  Realtek High Definition Audio Device
On Board Device 4 Information
	Type: Other
	Status: Enabled
	Description: CIR Device
On Board Device 5 Information
	Type: Other
	Status: Enabled
	Description: SD
On Board Device 6 Information
	Type: Other
	Status: Enabled
	Description: Intel Dual Band Wireless-AC 8265
On Board Device 7 Information
	Type: Other
	Status: Enabled
	Description: Bluetooth
On Board Device 8 Information
	Type: Other
	Status: Disabled
	Description: Thunderbolt
`
)

var biosInfoTests = map[string]string{
	"Vendor":          "LENOVO",
	"Version":         "29CN40WW(V2.17)",
	"Release Date":    "04/13/2011",
	"ROM Size":        "2048 kB",
	"Characteristics": "",
	"BIOS Revision":   "1.40",
}
var sysInfoTests = map[string]string{
	"Manufacturer":  "LENOVO",
	"Product Name":  "20042",
	"Version":       "Lenovo G560",
	"Serial Number": "2677240001087",
	"UUID":          "CB3E6A50-A77B-E011-88E9-B870F4165734",
	"Wake-up Type":  "Power Switch",
	"SKU Number":    "Calpella_CRB",
	"Family":        "Intel_Mobile",
}

var sysConfigurationTests = map[string]string{
	"Option 1": "String1 for Type12 Equipment Manufacturer",
	"Option 2": "String2 for Type12 Equipment Manufacturer",
	"Option 3": "String3 for Type12 Equipment Manufacturer",
	"Option 4": "String4 for Type12 Equipment Manufacturer",
}

var sysEventLogTests = map[string]string{
	"Area Length":                    "0 bytes",
	"Header Start Offset":            "0x0000",
	"Data Start Offset":              "0x0000",
	"Access Method":                  "General-purpose non-volatile data functions",
	"Access Address":                 "0x0000",
	"Status":                         "Valid, Not Full",
	"Change Token":                   "0x12345678",
	"Header Format":                  "OEM-specific",
	"Supported Log Type Descriptors": "3",
	"Descriptor 1":                   "POST memory resize",
	"Data Format 1":                  "None",
	"Descriptor 2":                   "POST error",
	"Data Format 2":                  "POST results bitmap",
	"Descriptor 3":                   "Log area reset/cleared",
	"Data Format 3":                  "None",
}

var sysBootTests = map[string]string{
	"Status": "No errors detected",
}

var processorTests = map[string]string{
	"Socket Designation": "CPU",
	"Type":               "Central Processor",
	"Family":             "Core 2 Duo",
	"Manufacturer":       "Intel(R) Corporation",
	"ID":                 "55 06 02 00 FF FB EB BF",
	"Signature":          "Type 0, Family 6, Model 37, Stepping 5",
	"Flags":              "",
	"Version":            "Intel(R) Core(TM) i3 CPU       M 370  @ 2.40GHz",
	"Voltage":            "0.0 V",
	"External Clock":     "1066 MHz",
	"Max Speed":          "2400 MHz",
	"Current Speed":      "2399 MHz",
	"Status":             "Populated, Enabled",
	"Upgrade":            "ZIF Socket",
	"L1 Cache Handle":    "0x0030",
	"L2 Cache Handle":    "0x002F",
	"L3 Cache Handle":    "0x002D",
	"Serial Number":      "Not Specified",
	"Asset Tag":          "FFFF",
	"Part Number":        "Not Specified",
	"Core Count":         "2",
	"Core Enabled":       "2",
	"Thread Count":       "4",
	"Characteristics":    "",
}

var onBoardDevicesTests = map[string]DMISubSection{
	"On Board Device 1 Information": {
		"Description": PropertyData{
			Val: "Intel(R) HD Graphics Device",
		},
		"Status": PropertyData{
			Val: "Enabled",
		},
		"Type": PropertyData{
			Val: "Video",
		},
	},
	"On Board Device 2 Information": {
		"Description": PropertyData{
			Val: "Intel(R) I219-V Gigabit Network Device",
		},
		"Status": PropertyData{
			Val: "Enabled",
		},
		"Type": PropertyData{
			Val: "Ethernet",
		},
	},
	"On Board Device 3 Information": {
		"Description": PropertyData{
			Val: "Realtek High Definition Audio Device",
		},
		"Status": PropertyData{
			Val: "Enabled",
		},
		"Type": PropertyData{
			Val: "Sound",
		},
	},
	"On Board Device 4 Information": {
		"Description": PropertyData{
			Val: "CIR Device",
		},
		"Status": PropertyData{
			Val: "Enabled",
		},
		"Type": PropertyData{
			Val: "Other",
		},
	},
	"On Board Device 5 Information": {
		"Description": PropertyData{
			Val: "SD",
		},
		"Status": PropertyData{
			Val: "Enabled",
		},
		"Type": PropertyData{
			Val: "Other",
		},
	},
	"On Board Device 6 Information": {
		"Description": PropertyData{
			Val: "Intel Dual Band Wireless-AC 8265",
		},
		"Status": PropertyData{
			Val: "Enabled",
		},
		"Type": PropertyData{
			Val: "Other",
		},
	},
	"On Board Device 7 Information": {
		"Description": PropertyData{
			Val: "Bluetooth",
		},
		"Status": PropertyData{
			Val: "Enabled",
		},
		"Type": PropertyData{
			Val: "Other",
		},
	},
	"On Board Device 8 Information": {
		"Description": PropertyData{
			Val: "Thunderbolt",
		},
		"Status": PropertyData{
			Val: "Disabled",
		},
		"Type": PropertyData{
			Val: "Other",
		},
	},
}

func TestParseSectionSimple(t *testing.T) {
	dmi, err := ParseDMI(sample1)
	if ok := assert.NoError(t, err); !ok {
		t.Fatal()
	}
	if ok := assert.Len(t, dmi, 1); !ok {
		t.Fatal()
	}
	if ok := assert.Len(t, dmi["System"].SubSections["System Information"], 8); !ok {
		t.Fatal()
	}

	for k, v := range sysInfoTests {
		if ok := assert.Equal(t, v, dmi["System"].SubSections["System Information"][k].Val); !ok {
			t.Fatal()
		}
	}

}
func TestParseSectionWithListProperty(t *testing.T) {
	dmi, err := ParseDMI(sample2)
	if ok := assert.NoError(t, err); !ok {
		t.Fatal()
	}

	if ok := assert.Len(t, dmi, 1); !ok {
		t.Fatal()
	}
	if ok := assert.Len(t, dmi["BIOS"].SubSections["BIOS Information"], 6); !ok {
		t.Fatal()
	}
	if ok := assert.Len(t, dmi["BIOS"].SubSections["BIOS Information"]["Characteristics"].Items, 18); !ok {
		t.Fatal()
	}

	for k, v := range biosInfoTests {
		if ok := assert.Equal(t, v, dmi["BIOS"].SubSections["BIOS Information"][k].Val); !ok {
			t.Fatal()
		}
	}

}

func TestParseMultipleSectionsSimple(t *testing.T) {
	dmi, err := ParseDMI(sample3)
	if ok := assert.NoError(t, err); !ok {
		t.Fatal()
	}
	if ok := assert.Len(t, dmi, 4); !ok {
		t.Fatal()
	}

	if ok := assert.Len(t, dmi["System"].SubSections["System Information"], 8); !ok {
		t.Fatal()
	}
	if ok := assert.Len(t, dmi["SystemEventLog"].SubSections["System Event Log"], 15); !ok {
		t.Fatal()
	}
	if ok := assert.Equal(t, DMITypeSystemBoot, dmi["SystemBoot"].Type); !ok {
		t.Fatal()
	}

	for k, v := range sysInfoTests {
		if ok := assert.Equal(t, v, dmi["System"].SubSections["System Information"][k].Val); !ok {
			t.Fatal()
		}
	}
	for k, v := range sysConfigurationTests {
		if ok := assert.Equal(t, v, dmi["SystemConfigurationOptions"].SubSections["System Configuration Options"][k].Val); !ok {
			t.Fatal()
		}
	}
	for k, v := range sysEventLogTests {
		if ok := assert.Equal(t, v, dmi["SystemEventLog"].SubSections["System Event Log"][k].Val); !ok {
			t.Fatal()
		}
	}
	for k, v := range sysBootTests {
		if ok := assert.Equal(t, v, dmi["SystemBoot"].SubSections["System Boot Information"][k].Val); !ok {
			t.Fatal()
		}
	}

}
func TestParseMultipleSectionsWithListProperties(t *testing.T) {
	dmi, err := ParseDMI(sample4)
	if ok := assert.NoError(t, err); !ok {
		t.Fatal()
	}
	if ok := assert.Len(t, dmi, 2); !ok {
		t.Fatal()
	}

	if ok := assert.Len(t, dmi["BIOS"].SubSections["BIOS Information"], 6); !ok {
		t.Fatal()
	}
	if ok := assert.Len(t, dmi["BIOS"].SubSections["BIOS Information"]["Characteristics"].Items, 18); !ok {
		t.Fatal()
	}

	if ok := assert.Len(t, dmi["Processor"].SubSections["Processor Information"], 24); !ok {
		t.Fatal()
	}

	if ok := assert.Len(t, dmi["Processor"].SubSections["Processor Information"]["Flags"].Items, 28); !ok {
		t.Fatal()
	}
	if ok := assert.Equal(t, "FPU (Floating-point unit on-chip)", dmi["Processor"].SubSections["Processor Information"]["Flags"].Items[0]); !ok {
		t.Fatal()
	}

	for k, v := range biosInfoTests {
		if ok := assert.Equal(t, v, dmi["BIOS"].SubSections["BIOS Information"][k].Val); !ok {
			t.Fatal()
		}
	}

	for k, v := range processorTests {
		if ok := assert.Equal(t, v, dmi["Processor"].SubSections["Processor Information"][k].Val); !ok {
			t.Fatal()
		}
	}
}
func TestParseTestOnBoardDevices(t *testing.T) {
	dmi, err := ParseDMI(sample5)
	if ok := assert.NoError(t, err); !ok {
		t.Fatal()
	}
	if ok := assert.Len(t, dmi, 1); !ok {
		t.Fatal()
	}
	if ok := assert.Len(t, onBoardDevicesTests, 8); !ok {
		t.Fatal()
	}
	if ok := assert.Len(t, dmi["OnBoardDevices"].SubSections, 8); !ok {
		t.Fatal()
	}
	if ok := assert.Equal(t, dmi["OnBoardDevices"].Type, DMITypeOnBoardDevices); !ok {
		t.Fatal()
	}
	for k, v := range onBoardDevicesTests {
		subSection, ok := dmi["OnBoardDevices"].SubSections[k]
		if !ok {
			t.Fatal()
		}
		for propertyName, property := range v {
			foundProperty, ok := subSection[propertyName]
			if !ok {
				t.Fatal()
			}
			if ok := assert.Equal(t, property.Val, foundProperty.Val); !ok {
				t.Fatal()
			}
			if ok := assert.Equal(t, len(property.Items), len(foundProperty.Items)); !ok {
				t.Fatal()
			}
			for index := range property.Items {
				if ok := assert.Equal(t, property.Items[index], foundProperty.Items[index]); !ok {
					t.Fatal()
				}
			}
		}
	}
}

func TestFullDmiDecodeOutputSample(t *testing.T) {
	_, err := ParseDMI(`# dmidecode 3.1\nGetting SMBIOS data from sysfs.\nSMBIOS 3.0.0 present.\nTable at 0x7AEAA000.\n\nHandle 0x0000, DMI type 0, 24 bytes\nBIOS Information\n	Vendor: Intel Corp.\n	Version: BNKBL357.86A.0062.2018.0222.1644\n	Release Date: 02/22/2018\n	Address: 0xF0000\n	Runtime Size: 64 kB\n	ROM Size: 8192 kB\n	Characteristics:\n		PCI is supported\n		BIOS is upgradeable\n		BIOS shadowing is allowed\n		Boot from CD is supported\n		Selectable boot is supported\n		BIOS ROM is socketed\n		EDD is supported\n		5.25"/1.2 MB floppy services are supported (int 13h)\n		3.5"/720 kB floppy services are supported (int 13h)\n		3.5"/2.88 MB floppy services are supported (int 13h)\n		Print screen service is supported (int 5h)\n		Serial services are supported (int 14h)\n		Printer services are supported (int 17h)\n		ACPI is supported\n		USB legacy is supported\n		BIOS boot specification is supported\n		Targeted content distribution is supported\n		UEFI is supported\n	BIOS Revision: 5.6\n	Firmware Revision: 8.12\n\nHandle 0x0001, DMI type 1, 27 bytes\nSystem Information\n	Manufacturer: Intel Corporation\n	Product Name: NUC7i5BNH\n	Version: J31169-311\n	Serial Number: G6BN81700BFU\n	UUID: BD11849D-5ADA-18C3-17BF-94C6911EC515\n	Wake-up Type: Power Switch\n	SKU Number:                                  \n	Family: Intel NUC\n\nHandle 0x0002, DMI type 2, 15 bytes\nBase Board Information\n	Manufacturer: Intel Corporation\n	Product Name: NUC7i5BNB\n	Version: J31144-310\n	Serial Number: GEBN816011HA\n	Asset Tag:                                  \n	Features:\n		Board is a hosting board\n		Board is replaceable\n	Location In Chassis: Default string\n	Chassis Handle: 0x0003\n	Type: Motherboard\n	Contained Object Handles: 0\n\nHandle 0x0003, DMI type 3, 22 bytes\nChassis Information\n	Manufacturer: Intel Corporation\n	Type: Desktop\n	Lock: Not Present\n	Version: 2\n	Serial Number:                                  \n	Asset Tag:                                  \n	Boot-up State: Safe\n	Power Supply State: Safe\n	Thermal State: Safe\n	Security Status: None\n	OEM Information: 0x00000000\n	Height: Unspecified\n	Number Of Power Cords: 1\n	Contained Elements: 0\n	SKU Number:                                  \n\nHandle 0x0004, DMI type 8, 9 bytes\nPort Connector Information\n	Internal Reference Designator: J3A1\n	Internal Connector Type: None\n	External Reference Designator: USB1\n	External Connector Type: Access Bus (USB)\n	Port Type: USB\n\nHandle 0x0005, DMI type 8, 9 bytes\nPort Connector Information\n	Internal Reference Designator: J3A1\n	Internal Connector Type: None\n	External Reference Designator: USB3\n	External Connector Type: Access Bus (USB)\n	Port Type: USB\n\nHandle 0x0006, DMI type 8, 9 bytes\nPort Connector Information\n	Internal Reference Designator: J5A1\n	Internal Connector Type: None\n	External Reference Designator: LAN\n	External Connector Type: RJ-45\n	Port Type: Network Port\n\nHandle 0x0007, DMI type 8, 9 bytes\nPort Connector Information\n	Internal Reference Designator: J5A1\n	Internal Connector Type: None\n	External Reference Designator: USB4\n	External Connector Type: Access Bus (USB)\n	Port Type: USB\n\nHandle 0x0008, DMI type 8, 9 bytes\nPort Connector Information\n	Internal Reference Designator: J5A1\n	Internal Connector Type: None\n	External Reference Designator: USB5\n	External Connector Type: Access Bus (USB)\n	Port Type: USB\n\nHandle 0x0009, DMI type 8, 9 bytes\nPort Connector Information\n	Internal Reference Designator: J9C1 - PCIE DOCKING CONN\n	Internal Connector Type: Other\n	External Reference Designator: Not Specified\n	External Connector Type: None\n	Port Type: Other\n\nHandle 0x000A, DMI type 8, 9 bytes\nPort Connector Information\n	Internal Reference Designator: J2B3 - CPU FAN\n	Internal Connector Type: Other\n	External Reference Designator: Not Specified\n	External Connector Type: None\n	Port Type: Other\n\nHandle 0x000B, DMI type 8, 9 bytes\nPort Connector Information\n	Internal Reference Designator: J6C2 - EXT HDMI\n	Internal Connector Type: Other\n	External Reference Designator: Not Specified\n	External Connector Type: None\n	Port Type: Other\n\nHandle 0x000C, DMI type 8, 9 bytes\nPort Connector Information\n	Internal Reference Designator: J2G1 - GFX VID\n	Internal Connector Type: Other\n	External Reference Designator: Not Specified\n	External Connector Type: None\n	Port Type: Other\n\nHandle 0x000D, DMI type 8, 9 bytes\nPort Connector Information\n	Internal Reference Designator: J1G6 - AC JACK\n	Internal Connector Type: Other\n	External Reference Designator: Not Specified\n	External Connector Type: None\n	Port Type: Other\n\nHandle 0x000E, DMI type 8, 9 bytes\nPort Connector Information\n	Internal Reference Designator: J7H2 - SATA PWR\n	Internal Connector Type: Other\n	External Reference Designator: Not Specified\n	External Connector Type: None\n	Port Type: Other\n\nHandle 0x000F, DMI type 9, 17 bytes\nSystem Slot Information\n	Designation: J6B2\n	Type: x16 PCI Express\n	Current Usage: In Use\n	Length: Long\n	ID: 0\n	Characteristics:\n		3.3 V is provided\n		Opening is shared\n		PME signal is supported\n	Bus Address: 0000:00:01.0\n\nHandle 0x0010, DMI type 9, 17 bytes\nSystem Slot Information\n	Designation: J6B1\n	Type: x1 PCI Express\n	Current Usage: In Use\n	Length: Short\n	ID: 1\n	Characteristics:\n		3.3 V is provided\n		Opening is shared\n		PME signal is supported\n	Bus Address: 0000:00:1c.3\n\nHandle 0x0011, DMI type 9, 17 bytes\nSystem Slot Information\n	Designation: J6D1\n	Type: x1 PCI Express\n	Current Usage: In Use\n	Length: Short\n	ID: 2\n	Characteristics:\n		3.3 V is provided\n		Opening is shared\n		PME signal is supported\n	Bus Address: 0000:00:1c.4\n\nHandle 0x0012, DMI type 9, 17 bytes\nSystem Slot Information\n	Designation: J7B1\n	Type: x1 PCI Express\n	Current Usage: In Use\n	Length: Short\n	ID: 3\n	Characteristics:\n		3.3 V is provided\n		Opening is shared\n		PME signal is supported\n	Bus Address: 0000:00:1c.5\n\nHandle 0x0013, DMI type 9, 17 bytes\nSystem Slot Information\n	Designation: J8B4\n	Type: x1 PCI Express\n	Current Usage: In Use\n	Length: Short\n	ID: 4\n	Characteristics:\n		3.3 V is provided\n		Opening is shared\n		PME signal is supported\n	Bus Address: 0000:00:1c.6\n\nHandle 0x0014, DMI type 10, 20 bytes\nOn Board Device 1 Information\n	Type: Video\n	Status: Enabled\n	Description:  Intel(R) HD Graphics Device\nOn Board Device 2 Information\n	Type: Ethernet\n	Status: Enabled\n	Description:  Intel(R) I219-V Gigabit Network Device\nOn Board Device 3 Information\n	Type: Sound\n	Status: Enabled\n	Description:  Realtek High Definition Audio Device\nOn Board Device 4 Information\n	Type: Other\n	Status: Enabled\n	Description: CIR Device\nOn Board Device 5 Information\n	Type: Other\n	Status: Enabled\n	Description: SD\nOn Board Device 6 Information\n	Type: Other\n	Status: Enabled\n	Description: Intel Dual Band Wireless-AC 8265\nOn Board Device 7 Information\n	Type: Other\n	Status: Enabled\n	Description: Bluetooth\nOn Board Device 8 Information\n	Type: Other\n	Status: Disabled\n	Description: Thunderbolt\n\nHandle 0x0015, DMI type 11, 5 bytes\nOEM Strings\n	String 1: Default string\n\nHandle 0x0016, DMI type 12, 5 bytes\nSystem Configuration Options\n	Option 1: Default string\n\nHandle 0x0017, DMI type 32, 20 bytes\nSystem Boot Information\n	Status: No errors detected\n\nHandle 0x0018, DMI type 34, 11 bytes\nManagement Device\n	Description: LM78-1\n	Type: LM78\n	Address: 0x00000000\n	Address Type: I/O Port\n\nHandle 0x0019, DMI type 26, 22 bytes\nVoltage Probe\n	Description: LM78A\n	Location: Motherboard\n	Status: OK\n	Maximum Value: Unknown\n	Minimum Value: Unknown\n	Resolution: Unknown\n	Tolerance: Unknown\n	Accuracy: Unknown\n	OEM-specific Information: 0x00000000\n	Nominal Value: Unknown\n\nHandle 0x001A, DMI type 36, 16 bytes\nManagement Device Threshold Data\n	Lower Non-critical Threshold: 1\n	Upper Non-critical Threshold: 2\n	Lower Critical Threshold: 3\n	Upper Critical Threshold: 4\n	Lower Non-recoverable Threshold: 5\n	Upper Non-recoverable Threshold: 6\n\nHandle 0x001B, DMI type 35, 11 bytes\nManagement Device Component\n	Description: Default string\n	Management Device Handle: 0x0018\n	Component Handle: 0x0019\n	Threshold Handle: 0x001A\n\nHandle 0x001C, DMI type 28, 22 bytes\nTemperature Probe\n	Description: LM78A\n	Location: Motherboard\n	Status: OK\n	Maximum Value: Unknown\n	Minimum Value: Unknown\n	Resolution: Unknown\n	Tolerance: Unknown\n	Accuracy: Unknown\n	OEM-specific Information: 0x00000000\n	Nominal Value: Unknown\n\nHandle 0x001D, DMI type 36, 16 bytes\nManagement Device Threshold Data\n	Lower Non-critical Threshold: 1\n	Upper Non-critical Threshold: 2\n	Lower Critical Threshold: 3\n	Upper Critical Threshold: 4\n	Lower Non-recoverable Threshold: 5\n	Upper Non-recoverable Threshold: 6\n\nHandle 0x001E, DMI type 35, 11 bytes\nManagement Device Component\n	Description: Default string\n	Management Device Handle: 0x0018\n	Component Handle: 0x001C\n	Threshold Handle: 0x001D\n\nHandle 0x001F, DMI type 27, 15 bytes\nCooling Device\n	Temperature Probe Handle: 0x001C\n	Type: Power Supply Fan\n	Status: OK\n	Cooling Unit Group: 1\n	OEM-specific Information: 0x00000000\n	Nominal Speed: Unknown Or Non-rotating\n	Description: Cooling Dev 1\n\nHandle 0x0020, DMI type 36, 16 bytes\nManagement Device Threshold Data\n	Lower Non-critical Threshold: 1\n	Upper Non-critical Threshold: 2\n	Lower Critical Threshold: 3\n	Upper Critical Threshold: 4\n	Lower Non-recoverable Threshold: 5\n	Upper Non-recoverable Threshold: 6\n\nHandle 0x0021, DMI type 35, 11 bytes\nManagement Device Component\n	Description: Default string\n	Management Device Handle: 0x0018\n	Component Handle: 0x001F\n	Threshold Handle: 0x0020\n\nHandle 0x0022, DMI type 27, 15 bytes\nCooling Device\n	Temperature Probe Handle: 0x001C\n	Type: Power Supply Fan\n	Status: OK\n	Cooling Unit Group: 1\n	OEM-specific Information: 0x00000000\n	Nominal Speed: Unknown Or Non-rotating\n	Description: Not Specified\n\nHandle 0x0023, DMI type 36, 16 bytes\nManagement Device Threshold Data\n	Lower Non-critical Threshold: 1\n	Upper Non-critical Threshold: 2\n	Lower Critical Threshold: 3\n	Upper Critical Threshold: 4\n	Lower Non-recoverable Threshold: 5\n	Upper Non-recoverable Threshold: 6\n\nHandle 0x0024, DMI type 35, 11 bytes\nManagement Device Component\n	Description: Default string\n	Management Device Handle: 0x0018\n	Component Handle: 0x0022\n	Threshold Handle: 0x0023\n\nHandle 0x0025, DMI type 29, 22 bytes\nElectrical Current Probe\n	Description: ABC\n	Location: Motherboard\n	Status: OK\n	Maximum Value: Unknown\n	Minimum Value: Unknown\n	Resolution: Unknown\n	Tolerance: Unknown\n	Accuracy: Unknown\n	OEM-specific Information: 0x00000000\n	Nominal Value: Unknown\n\nHandle 0x0026, DMI type 36, 16 bytes\nManagement Device Threshold Data\n\nHandle 0x0027, DMI type 35, 11 bytes\nManagement Device Component\n	Description: Default string\n	Management Device Handle: 0x0018\n	Component Handle: 0x0025\n	Threshold Handle: 0x0026\n\nHandle 0x0028, DMI type 26, 22 bytes\nVoltage Probe\n	Description: LM78A\n	Location: Power Unit\n	Status: OK\n	Maximum Value: Unknown\n	Minimum Value: Unknown\n	Resolution: Unknown\n	Tolerance: Unknown\n	Accuracy: Unknown\n	OEM-specific Information: 0x00000000\n	Nominal Value: Unknown\n\nHandle 0x0029, DMI type 28, 22 bytes\nTemperature Probe\n	Description: LM78A\n	Location: Power Unit\n	Status: OK\n	Maximum Value: Unknown\n	Minimum Value: Unknown\n	Resolution: Unknown\n	Tolerance: Unknown\n	Accuracy: Unknown\n	OEM-specific Information: 0x00000000\n	Nominal Value: Unknown\n\nHandle 0x002A, DMI type 27, 15 bytes\nCooling Device\n	Temperature Probe Handle: 0x0029\n	Type: Power Supply Fan\n	Status: OK\n	Cooling Unit Group: 1\n	OEM-specific Information: 0x00000000\n	Nominal Speed: Unknown Or Non-rotating\n	Description: Cooling Dev 1\n\nHandle 0x002B, DMI type 29, 22 bytes\nElectrical Current Probe\n	Description: ABC\n	Location: Power Unit\n	Status: OK\n	Maximum Value: Unknown\n	Minimum Value: Unknown\n	Resolution: Unknown\n	Tolerance: Unknown\n	Accuracy: Unknown\n	OEM-specific Information: 0x00000000\n	Nominal Value: Unknown\n\nHandle 0x002C, DMI type 39, 22 bytes\nSystem Power Supply\n	Power Unit Group: 1\n	Location: To Be Filled By O.E.M.\n	Name: To Be Filled By O.E.M.\n	Manufacturer: To Be Filled By O.E.M.\n	Serial Number: To Be Filled By O.E.M.\n	Asset Tag: To Be Filled By O.E.M.\n	Model Part Number: To Be Filled By O.E.M.\n	Revision: To Be Filled By O.E.M.\n	Max Power Capacity: Unknown\n	Status: Present, OK\n	Type: Switching\n	Input Voltage Range Switching: Auto-switch\n	Plugged: Yes\n	Hot Replaceable: No\n	Input Voltage Probe Handle: 0x0028\n	Cooling Device Handle: 0x002A\n	Input Current Probe Handle: 0x002B\n\nHandle 0x002D, DMI type 41, 11 bytes\nOnboard Device\n	Reference Designation:  CPU\n	Type: Video\n	Status: Enabled\n	Type Instance: 1\n	Bus Address: 0000:00:02.0\n\nHandle 0x002E, DMI type 41, 11 bytes\nOnboard Device\n	Reference Designation:  LAN\n	Type: Ethernet\n	Status: Enabled\n	Type Instance: 1\n	Bus Address: 0000:00:1f.6\n\nHandle 0x002F, DMI type 41, 11 bytes\nOnboard Device\n	Reference Designation:  AUDIO\n	Type: Sound\n	Status: Enabled\n	Type Instance: 1\n	Bus Address: 00ff:00:1f.7\n\nHandle 0x0030, DMI type 41, 11 bytes\nOnboard Device\n	Reference Designation:  CIR Device\n	Type: Other\n	Status: Enabled\n	Type Instance: 1\n	Bus Address: 00ff:00:1f.7\n\nHandle 0x0031, DMI type 41, 11 bytes\nOnboard Device\n	Reference Designation:  SD\n	Type: Other\n	Status: Enabled\n	Type Instance: 1\n	Bus Address: 0000:00:00.0\n\nHandle 0x0032, DMI type 41, 11 bytes\nOnboard Device\n	Reference Designation:  Intel Dual Band\n	Type: Other\n	Status: Enabled\n	Type Instance: 1\n	Bus Address: 0000:00:00.0\n\nHandle 0x0033, DMI type 41, 11 bytes\nOnboard Device\n	Reference Designation:  Bluetooth\n	Type: Other\n	Status: Enabled\n	Type Instance: 1\n	Bus Address: 00ff:00:1f.7\n\nHandle 0x0034, DMI type 41, 11 bytes\nOnboard Device\n	Reference Designation:  Thunderbolt\n	Type: Other\n	Status: Disabled\n	Type Instance: 1\n	Bus Address: 00ff:00:1f.7\n\nHandle 0x0035, DMI type 16, 23 bytes\nPhysical Memory Array\n	Location: System Board Or Motherboard\n	Use: System Memory\n	Error Correction Type: None\n	Maximum Capacity: 32 GB\n	Error Information Handle: Not Provided\n	Number Of Devices: 2\n\nHandle 0x0036, DMI type 17, 40 bytes\nMemory Device\n	Array Handle: 0x0035\n	Error Information Handle: Not Provided\n	Total Width: 64 bits\n	Data Width: 64 bits\n	Size: 8192 MB\n	Form Factor: SODIMM\n	Set: None\n	Locator: ChannelA-DIMM0\n	Bank Locator: BANK 0\n	Type: DDR4\n	Type Detail: Synchronous Unbuffered (Unregistered)\n	Speed: 2400 MT/s\n	Manufacturer: 859B\n	Serial Number: E0AD159C\n	Asset Tag: 9876543210\n	Part Number: CT8G4SFD824A.M16FB  \n	Rank: 2\n	Configured Clock Speed: 2133 MT/s\n	Minimum Voltage: 1.2 V\n	Maximum Voltage: 1.2 V\n	Configured Voltage: 1.2 V\n\nHandle 0x0037, DMI type 17, 40 bytes\nMemory Device\n	Array Handle: 0x0035\n	Error Information Handle: Not Provided\n	Total Width: 64 bits\n	Data Width: 64 bits\n	Size: 8192 MB\n	Form Factor: SODIMM\n	Set: None\n	Locator: ChannelB-DIMM0\n	Bank Locator: BANK 2\n	Type: DDR4\n	Type Detail: Synchronous Unbuffered (Unregistered)\n	Speed: 2400 MT/s\n	Manufacturer: 859B\n	Serial Number: E0AD159D\n	Asset Tag: 9876543210\n	Part Number: CT8G4SFD824A.M16FB  \n	Rank: 2\n	Configured Clock Speed: 2133 MT/s\n	Minimum Voltage: 1.2 V\n	Maximum Voltage: 1.2 V\n	Configured Voltage: 1.2 V\n\nHandle 0x0038, DMI type 19, 31 bytes\nMemory Array Mapped Address\n	Starting Address: 0x00000000000\n	Ending Address: 0x003FFFFFFFF\n	Range Size: 16 GB\n	Physical Array Handle: 0x0035\n	Partition Width: 2\n\nHandle 0x0039, DMI type 43, 31 bytes\nTPM Device\n	Vendor ID: CTNI\n	Specification Version: 2.0	Firmware Revision: 11.1\n	Description: INTEL	Characteristics:\n		Family configurable via platform software support\n	OEM-specific Information: 0x00000000\n\nHandle 0x003A, DMI type 7, 19 bytes\nCache Information\n	Socket Designation: L1 Cache\n	Configuration: Enabled, Not Socketed, Level 1\n	Operational Mode: Write Back\n	Location: Internal\n	Installed Size: 128 kB\n	Maximum Size: 128 kB\n	Supported SRAM Types:\n		Synchronous\n	Installed SRAM Type: Synchronous\n	Speed: Unknown\n	Error Correction Type: Parity\n	System Type: Unified\n	Associativity: 8-way Set-associative\n\nHandle 0x003B, DMI type 7, 19 bytes\nCache Information\n	Socket Designation: L2 Cache\n	Configuration: Enabled, Not Socketed, Level 2\n	Operational Mode: Write Back\n	Location: Internal\n	Installed Size: 512 kB\n	Maximum Size: 512 kB\n	Supported SRAM Types:\n		Synchronous\n	Installed SRAM Type: Synchronous\n	Speed: Unknown\n	Error Correction Type: Single-bit ECC\n	System Type: Unified\n	Associativity: 4-way Set-associative\n\nHandle 0x003C, DMI type 7, 19 bytes\nCache Information\n	Socket Designation: L3 Cache\n	Configuration: Enabled, Not Socketed, Level 3\n	Operational Mode: Write Back\n	Location: Internal\n	Installed Size: 4096 kB\n	Maximum Size: 4096 kB\n	Supported SRAM Types:\n		Synchronous\n	Installed SRAM Type: Synchronous\n	Speed: Unknown\n	Error Correction Type: Multi-bit ECC\n	System Type: Unified\n	Associativity: 16-way Set-associative\n\nHandle 0x003D, DMI type 4, 48 bytes\nProcessor Information\n	Socket Designation: U3E1\n	Type: Central Processor\n	Family: Core i5\n	Manufacturer: Intel(R) Corporation\n	ID: E9 06 08 00 FF FB EB BF\n	Signature: Type 0, Family 6, Model 142, Stepping 9\n	Flags:\n		FPU (Floating-point unit on-chip)\n		VME (Virtual mode extension)\n		DE (Debugging extension)\n		PSE (Page size extension)\n		TSC (Time stamp counter)\n		MSR (Model specific registers)\n		PAE (Physical address extension)\n		MCE (Machine check exception)\n		CX8 (CMPXCHG8 instruction supported)\n		APIC (On-chip APIC hardware supported)\n		SEP (Fast system call)\n		MTRR (Memory type range registers)\n		PGE (Page global enable)\n		MCA (Machine check architecture)\n		CMOV (Conditional move instruction supported)\n		PAT (Page attribute table)\n		PSE-36 (36-bit page size extension)\n		CLFSH (CLFLUSH instruction supported)\n		DS (Debug store)\n		ACPI (ACPI supported)\n		MMX (MMX technology supported)\n		FXSR (FXSAVE and FXSTOR instructions supported)\n		SSE (Streaming SIMD extensions)\n		SSE2 (Streaming SIMD extensions 2)\n		SS (Self-snoop)\n		HTT (Multi-threading)\n		TM (Thermal monitor supported)\n		PBE (Pending break enabled)\n	Version: Intel(R) Core(TM) i5-7260U CPU @ 2.20GHz\n	Voltage: 0.8 V\n	External Clock: 100 MHz\n	Max Speed: 2400 MHz\n	Current Speed: 2200 MHz\n	Status: Populated, Enabled\n	Upgrade: Socket BGA1356\n	L1 Cache Handle: 0x003A\n	L2 Cache Handle: 0x003B\n	L3 Cache Handle: 0x003C\n	Serial Number: To Be Filled By O.E.M.\n	Asset Tag: To Be Filled By O.E.M.\n	Part Number: To Be Filled By O.E.M.\n	Core Count: 2\n	Core Enabled: 2\n	Thread Count: 4\n	Characteristics:\n		64-bit capable\n		Multi-Core\n		Hardware Thread\n		Execute Protection\n		Enhanced Virtualization\n		Power/Performance Control\n\nHandle 0x003E, DMI type 20, 35 bytes\nMemory Device Mapped Address\n	Starting Address: 0x00000000000\n	Ending Address: 0x001FFFFFFFF\n	Range Size: 8 GB\n	Physical Device Handle: 0x0036\n	Memory Array Mapped Address Handle: 0x0038\n	Partition Row Position: Unknown\n	Interleave Position: 1\n	Interleaved Data Depth: 1\n\nHandle 0x003F, DMI type 20, 35 bytes\nMemory Device Mapped Address\n	Starting Address: 0x00200000000\n	Ending Address: 0x003FFFFFFFF\n	Range Size: 8 GB\n	Physical Device Handle: 0x0037\n	Memory Array Mapped Address Handle: 0x0038\n	Partition Row Position: Unknown\n	Interleave Position: 2\n	Interleaved Data Depth: 1\n\nHandle 0x0040, DMI type 130, 20 bytes\nOEM-specific Type\n	Header and Data:\n		82 14 40 00 24 41 4D 54 00 00 00 00 00 A5 AF 02\n		C0 00 00 00\n\nHandle 0x0041, DMI type 131, 64 bytes\nOEM-specific Type\n	Header and Data:\n		83 40 41 00 31 00 00 00 00 00 00 00 00 00 00 00\n		F8 00 4E 9D 00 00 00 00 01 00 00 00 08 00 0B 00\n		61 0D 32 00 00 00 00 00 FE 00 D8 15 00 00 00 00\n		00 00 00 00 22 00 00 00 76 50 72 6F 00 00 00 00\n\nHandle 0x0042, DMI type 221, 33 bytes\nOEM-specific Type\n	Header and Data:\n		DD 21 42 00 04 01 00 02 08 01 00 00 02 00 00 00\n		00 84 00 03 00 00 05 00 00 00 04 00 FF FF FF FF\n		FF\n	Strings:\n		Reference Code - CPU\n		uCode Version\n		TXT ACM Version\n		BIOS Guard Version\n\nHandle 0x0043, DMI type 221, 26 bytes\nOEM-specific Type\n	Header and Data:\n		DD 1A 43 00 03 01 00 02 08 01 00 00 02 00 00 00\n		00 00 00 03 04 0B 08 32 61 0D\n	Strings:\n		Reference Code - ME 11.0\n		MEBx version\n		ME Firmware Version\n		Consumer SKU\n\nHandle 0x0044, DMI type 221, 75 bytes\nOEM-specific Type\n	Header and Data:\n		DD 4B 44 00 0A 01 00 02 08 01 00 00 02 03 FF FF\n		FF FF FF 04 00 FF FF FF 21 00 05 00 FF FF FF 21\n		00 06 00 FF FF FF FF FF 07 00 3E 00 00 00 00 08\n		00 34 00 00 00 00 09 00 0B 00 00 00 00 0A 00 3E\n		00 00 00 00 0B 00 34 00 00 00 00\n	Strings:\n		Reference Code - SKL PCH\n		PCH-CRID Status\n		Disabled\n		PCH-CRID Original Value\n		PCH-CRID New Value\n		OPROM - RST - RAID\n		SKL PCH H Bx Hsio Version\n		SKL PCH H Dx Hsio Version\n		KBL PCH H Ax Hsio Version\n		SKL PCH LP Bx Hsio Version\n		SKL PCH LP Cx Hsio Version\n\nHandle 0x0045, DMI type 221, 54 bytes\nOEM-specific Type\n	Header and Data:\n		DD 36 45 00 07 01 00 02 08 01 00 00 02 00 02 08\n		01 00 00 03 00 02 08 01 00 00 04 05 FF FF FF FF\n		FF 06 00 FF FF FF 03 00 07 00 FF FF FF 03 00 08\n		00 FF FF FF FF FF\n	Strings:\n		Reference Code - SA - System Agent\n		Reference Code - MRC\n		SA - PCIe Version\n		SA-CRID Status\n		Disabled\n		SA-CRID Original Value\n		SA-CRID New Value\n		OPROM - VBIOS\n\nHandle 0x0046, DMI type 221, 96 bytes\nOEM-specific Type\n	Header and Data:\n		DD 60 46 00 0D 01 00 00 00 00 A6 00 02 00 FF FF\n		FF FF FF 03 04 FF FF FF FF FF 05 06 FF FF FF FF\n		FF 07 08 FF FF FF FF FF 09 00 00 00 00 00 00 0A\n		00 FF FF FF FF 00 0B 00 08 0C 00 00 00 0C 00 00\n		09 00 66 10 0D 00 FF FF FF FF FF 0E 00 FF FF FF\n		FF FF 0F 10 01 03 04 01 01 11 00 00 07 03 00 00\n	Strings:\n		Lan Phy Version\n		Sensor Firmware Version\n		Debug Mode Status\n		Disabled\n		Performance Mode Status\n		Disabled\n		Debug Use USB(Disabled:Serial)\n		Disabled\n		ICC Overclocking Version\n		UNDI Version\n		EC FW Version\n		GOP Version\n		Base EC FW Version\n		EC-EC Protocol Version\n		Royal Park Version\n		BP1.3.4.1_RP01\n		Platform Version\n\nHandle 0x0047, DMI type 136, 6 bytes\nOEM-specific Type\n	Header and Data:\n		88 06 47 00 00 00\n\nHandle 0x0048, DMI type 14, 20 bytes\nGroup Associations\n	Name: Firmware Version Info\n	Items: 5\n		0x0042 (OEM-specific)\n		0x0043 (OEM-specific)\n		0x0044 (OEM-specific)\n		0x0045 (OEM-specific)\n		0x0046 (OEM-specific)\n\nHandle 0x0049, DMI type 14, 8 bytes\nGroup Associations\n	Name: $MEI\n	Items: 1\n		0x0000 (OEM-specific)\n\nHandle 0x004A, DMI type 219, 81 bytes\nOEM-specific Type\n	Header and Data:\n		DB 51 4A 00 01 03 01 45 02 00 90 06 01 00 66 20\n		00 00 00 00 40 08 00 00 00 00 00 00 00 00 40 02\n		FF FF FF FF FF FF FF FF FF FF FF FF FF FF FF FF\n		FF FF FF FF FF FF FF FF 03 00 00 00 80 00 00 00\n		00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00\n		00\n	Strings:\n		MEI1\n		MEI2\n		MEI3\n\nHandle 0x004B, DMI type 127, 4 bytes\nEnd Of Table\n\n`)
	if ok := assert.NoError(t, err); !ok {
		t.Fatal()
	}
}

func BenchmarkParseMultipleSectionsWithLists(b *testing.B) {
	// run the Fib function b.N times
	for n := 0; n < b.N; n++ {
		ParseDMI(sample4)
	}
}
