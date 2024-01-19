package cip1852

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

type Segment struct {
	Value    uint64
	IsHarden bool
}

func (seg Segment) ToString() string {
	var harden = ""
	if seg.IsHarden {
		harden = "'"
	}
	return fmt.Sprint(seg.Value, harden)
}

func stringToSegment(pathSeg string) (*Segment, error) {
	var value = strings.ReplaceAll(pathSeg, "'", "")
	var isHarden = strings.Contains(pathSeg, "'")
	parseInt, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return nil, errors.New("path error")
	}
	return &Segment{
		uint64(parseInt), isHarden,
	}, nil
}

type DerivationPath struct {
	Purpose  Segment
	CoinType Segment
	Account  Segment
	Role     Segment
	Index    Segment
}

func (path DerivationPath) ToString() string {
	return fmt.Sprint("m", "/", path.Purpose.ToString(), "/", path.CoinType.ToString(), "/", path.Account.ToString(), "/", path.Role.ToString(), "/", path.Index.ToString())
}

func CreateExternalAddressPath(index uint64) DerivationPath {
	return CreateExternalAddressPathForAccount(index, 0)
}

func CreateExternalAddressPathForAccount(index, account uint64) DerivationPath {
	return DerivationPath{
		Purpose: Segment{
			Value:    1852,
			IsHarden: true,
		},
		CoinType: Segment{
			Value:    1815,
			IsHarden: true,
		},
		Account: Segment{
			Value:    account,
			IsHarden: true,
		},
		Role: Segment{
			Value:    0,
			IsHarden: false,
		},
		Index: Segment{
			Value:    index,
			IsHarden: false,
		},
	}
}

func CreateFromPath(path string) (*DerivationPath, error) {
	if path == "" {
		return nil, errors.New("path is nil")
	}
	split := strings.Split(path, "/")

	if len(split) != 6 {
		return nil, errors.New("path format error")
	}

	purpos, err := stringToSegment(split[1])
	if err != nil {
		return nil, err
	}
	coinType, err := stringToSegment(split[2])
	if err != nil {
		return nil, err
	}
	account, err := stringToSegment(split[3])
	if err != nil {
		return nil, err
	}
	role, err := stringToSegment(split[4])
	if err != nil {
		return nil, err
	}
	index, err := stringToSegment(split[5])
	if err != nil {
		return nil, err
	}

	return &DerivationPath{
		Purpose:  *purpos,
		CoinType: *coinType,
		Account:  *account,
		Role:     *role,
		Index:    *index,
	}, nil
}

func CreateInternalAddressPath(index, acount uint64) DerivationPath {
	return DerivationPath{
		Purpose: Segment{
			Value:    1852,
			IsHarden: true,
		},
		CoinType: Segment{
			Value:    1815,
			IsHarden: true,
		},
		Account: Segment{
			Value:    acount,
			IsHarden: true,
		},
		Role: Segment{
			Value:    1,
			IsHarden: false,
		},
		Index: Segment{
			Value:    index,
			IsHarden: false,
		},
	}
}

func CreateStakeAddressPath(index, acount uint64) DerivationPath {
	return DerivationPath{
		Purpose: Segment{
			Value:    1852,
			IsHarden: true,
		},
		CoinType: Segment{
			Value:    1815,
			IsHarden: true,
		},
		Account: Segment{
			Value:    acount,
			IsHarden: true,
		},
		Role: Segment{
			Value:    2,
			IsHarden: false,
		},
		Index: Segment{
			Value:    index,
			IsHarden: false,
		},
	}
}
