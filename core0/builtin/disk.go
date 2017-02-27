package builtin

import (
	"encoding/json"
	"fmt"
	"github.com/g8os/core0/base/pm"
	"github.com/g8os/core0/base/pm/core"
	"github.com/g8os/core0/base/pm/process"
	"github.com/pborman/uuid"
	"io/ioutil"
	"regexp"
	"strconv"
)

/*
Implementation for disk.getinfo
Note that, the rest of the disk extension implementation is done in conf/disk.toml file
*/

var (
	freeSpaceRegex = regexp.MustCompile(`(?m:^\s*(\d+)B\s+(\d+)B\s+(\d+)B\s+Free Space$)`)
)

type diskMgr struct{}

func init() {
	d := (*diskMgr)(nil)
	pm.CmdMap["disk.getinfo"] = process.NewInternalProcessFactory(d.info)
}

type diskInfo struct {
	Disk string `json:"disk"`
	Part string `json:"part"`
}

type DiskFreeBlock struct {
	Start uint64 `json:"start"`
	End   uint64 `json:"end"`
	Size  uint64 `json:"size"`
}

type DiskInfoResult struct {
	Start     uint64          `json:"start"`
	End       uint64          `json:"end"`
	Size      uint64          `json:"size"`
	BlockSize uint64          `json:"blocksize"`
	Free      []DiskFreeBlock `json:"free"`
}

func (d *diskMgr) readUInt64(p string) (uint64, error) {
	bytes, err := ioutil.ReadFile(p)
	if err != nil {
		return 0, err
	}
	var r uint64
	if _, err := fmt.Sscanf(string(bytes), "%d", &r); err != nil {
		return 0, err
	}

	return r, nil
}

func (d *diskMgr) blockSize(dev string) (uint64, error) {
	runner, err := pm.GetManager().RunCmd(&core.Command{
		ID:      uuid.New(),
		Command: process.CommandSystem,
		Arguments: core.MustArguments(process.SystemCommandArguments{
			Name: "lsblk",
			Args: []string{"-n", "-r", "-o", "PHY-SEC", fmt.Sprintf("/dev/%s", dev)},
		}),
	})

	if err != nil {
		return 0, err
	}
	result := runner.Wait()

	if result.State != core.StateSuccess {
		return 0, fmt.Errorf("failed to run lsbl: %v", result.Streams)
	}

	stdout := ""
	if len(result.Streams) > 1 {
		stdout = result.Streams[0]
	}

	var bs uint64
	if _, err := fmt.Sscanf(stdout, "%d", &bs); err != nil {
		return 0, err
	}

	return bs, nil
}

func (d *diskMgr) getFreeBlocks(disk string) ([]DiskFreeBlock, error) {
	blocks := make([]DiskFreeBlock, 0)
	runner, err := pm.GetManager().RunCmd(&core.Command{
		ID:      uuid.New(),
		Command: process.CommandSystem,
		Arguments: core.MustArguments(process.SystemCommandArguments{
			Name: "parted",
			Args: []string{fmt.Sprintf("/dev/%s", disk), "unit", "B", "print", "free"},
		}),
	})

	if err != nil {
		return blocks, err
	}

	result := runner.Wait()
	if result.State != core.StateSuccess {
		return blocks, fmt.Errorf("failed to run parted: %v", result.Streams)
	}

	stdout := ""

	if len(result.Streams) > 1 {
		stdout = result.Streams[0]
	}

	matches := freeSpaceRegex.FindAllStringSubmatch(stdout, -1)
	for _, match := range matches {
		bstart, _ := strconv.ParseUint(match[1], 10, 64)
		bend, _ := strconv.ParseUint(match[2], 10, 64)
		bsize, _ := strconv.ParseUint(match[3], 10, 64)

		blocks = append(blocks, DiskFreeBlock{
			Start: bstart,
			End:   bend,
			Size:  bsize,
		})
	}

	return blocks, nil
}

func (d *diskMgr) diskInfo(disk string) (*DiskInfoResult, error) {
	var info DiskInfoResult

	bs, err := d.blockSize(disk)
	if err != nil {
		return nil, err
	}

	size, err := d.readUInt64(fmt.Sprintf("/sys/block/%s/size", disk))
	if err != nil {
		return nil, err
	}
	info.Size = size * bs
	info.End = (size * bs) - 1

	info.BlockSize = bs
	//get free blocks.
	blocks, err := d.getFreeBlocks(disk)
	if err != nil {
		return nil, err
	}

	info.Free = blocks

	return &info, nil
}

func (d *diskMgr) partInfo(disk, part string) (*DiskInfoResult, error) {
	var info DiskInfoResult
	bs, err := d.blockSize(part)
	if err != nil {
		return nil, err
	}

	start, err := d.readUInt64(fmt.Sprintf("/sys/block/%s/%s/start", disk, part))
	if err != nil {
		return nil, err
	}

	size, err := d.readUInt64(fmt.Sprintf("/sys/block/%s/%s/size", disk, part))
	if err != nil {
		return nil, err
	}

	info.Start = start * bs
	info.Size = size * bs
	info.End = info.Start + info.Size - 1

	info.BlockSize = bs
	info.Free = make([]DiskFreeBlock, 0) //this is just to make the return consistent
	return &info, nil
}

func (d *diskMgr) info(cmd *core.Command) (interface{}, error) {
	var args diskInfo

	if err := json.Unmarshal(*cmd.Arguments, &args); err != nil {
		return nil, err
	}

	if args.Part == "" {
		return d.diskInfo(args.Disk)
	}

	return d.partInfo(args.Disk, args.Part)
}
