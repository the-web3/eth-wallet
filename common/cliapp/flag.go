package cliapp

import (
	"fmt"

	"github.com/urfave/cli/v2"
)

type CloneableGeneric interface {
	cli.Generic
	Clone() any
}

func ProtectFlags(flags []cli.Flag) []cli.Flag {
	out := make([]cli.Flag, 0, len(flags))
	for _, f := range flags {
		fCopy, err := cloneFlag(f)
		if err != nil {
			panic(fmt.Errorf("failed to clone flag %q: %w", f.Names()[0], err))
		}
		out = append(out, fCopy)
	}
	return out
}

func cloneFlag(f cli.Flag) (cli.Flag, error) {
	switch typedFlag := f.(type) {
	case *cli.GenericFlag:
		if genValue, ok := typedFlag.Value.(CloneableGeneric); ok {
			cpy := *typedFlag
			cpyVal, ok := genValue.Clone().(cli.Generic)
			if !ok {
				return nil, fmt.Errorf("cloned Generic value is not Generic: %T", typedFlag)
			}
			cpy.Value = cpyVal
			return &cpy, nil
		} else {
			return nil, fmt.Errorf("cannot clone Generic value: %T", typedFlag)
		}
	default:
		return f, nil
	}
}
