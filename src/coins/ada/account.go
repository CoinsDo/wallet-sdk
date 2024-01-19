package ada

import (
	"errors"
	"github.com/btcsuite/btcd/btcutil/base58"
	"github.com/echovl/bech32"
	"strings"
)

type AddressType int

const (
	Address_Type_Base       = 1
	Address_Type_Ptr        = 2
	Address_Type_Enterprise = 3
	Address_Type_Reward     = 4
	Address_Type_Byron      = 5
)

type HeaderKind byte

const (
	HeaderKind_Base       = 0b0000_0000
	HeaderKind_Ptr        = 0b0100_0000
	HeaderKind_Enterprise = 0b0110_0000
	HeaderKind_Reward     = 0b1110_0000
	HeaderKind_Byron      = 0b1000_0000
)

func GetEnterpriseAddress(paymentKeyHash []byte, network Network) (string, error) {
	return GetAddress(paymentKeyHash, nil, HeaderKind_Enterprise, network, Address_Type_Enterprise)
}

func GetBaseAddress(paymentKeyHash, stakeKeyhash []byte, network Network) (string, error) {
	return GetAddress(paymentKeyHash, stakeKeyhash, HeaderKind_Base, network, Address_Type_Base)
}

func GetByronAddress(paymentKeyHash, stakeKeyhash []byte, network Network) (string, error) {
	return GetAddress(paymentKeyHash, stakeKeyhash, HeaderKind_Byron, network, Address_Type_Byron)
}

func GetRewardAddress(stakeKeyhash []byte, network Network) (string, error) {
	return GetAddress(nil, stakeKeyhash, HeaderKind_Reward, network, Address_Type_Reward)
}

func GetAddress(paymentKeyHash, stakeKeyHash []byte, headerKind byte, network Network, addressType AddressType) (string, error) {

	//networkId := network.NetworkId

	addrPrefix := getPrefixHeader(addressType) + getPrefixTail(network)

	header := getAddressHeader(headerKind, network)

	var addressArray []byte
	switch addressType {
	case Address_Type_Base:
		addressArray = make([]byte, len(paymentKeyHash)+len(stakeKeyHash)+1)
		addressArray[0] = header
		copy(addressArray[1:], paymentKeyHash)
		copy(addressArray[1+len(paymentKeyHash):], stakeKeyHash)

		break
	case Address_Type_Ptr:
		addressArray = make([]byte, len(paymentKeyHash)+len(stakeKeyHash)+1)
		addressArray[0] = header
		copy(addressArray[1:], paymentKeyHash)
		copy(addressArray[1+len(paymentKeyHash):], stakeKeyHash)

		break
	case Address_Type_Enterprise:
		addressArray = make([]byte, len(paymentKeyHash)+1)
		addressArray[0] = header
		copy(addressArray[1:], paymentKeyHash)
		break
	case Address_Type_Reward:
		addressArray = make([]byte, len(stakeKeyHash)+1)
		addressArray[0] = header
		copy(addressArray[1:], stakeKeyHash)
		break
	case Address_Type_Byron:
		addressArray = make([]byte, len(paymentKeyHash)+len(stakeKeyHash)+1)
		addressArray[0] = header
		copy(addressArray[1:], paymentKeyHash)
		copy(addressArray[1+len(paymentKeyHash):], stakeKeyHash)

		return base58.Encode(addressArray), nil

	}
	return encodeSegWitAddress(addrPrefix, []byte("1")[0], addressArray)

}

func encodeSegWitAddress(hrp string, witnessVersion byte, witnessProgram []byte) (string, error) {
	// Group the address bytes into 5 bit groups, as this is what is used to
	// encode each character in the address string.
	converted, err := ConvertBits(witnessProgram, 8, 5, true)
	if err != nil {
		return "", err
	}
	builder := strings.Builder{}
	checksum := writeBech32Checksum(hrp, converted)

	b32Arr := make([]byte, len(converted)+len(checksum))
	copy(b32Arr, converted[:])
	copy(b32Arr[len(converted):], checksum[:])
	builder.WriteString(hrp)
	builder.WriteString("1")
	for _, by := range b32Arr {
		builder.WriteByte(charset[by])
	}
	return builder.String(), nil
}

func ConvertBits(data []byte, fromBits, toBits uint8, pad bool) ([]byte, error) {
	acc := 0
	bits := 0
	maxv := (1 << toBits) - 1
	maxacc := (1 << (fromBits + toBits - 1)) - 1
	var results []byte
	for _, b := range data {

		// Discard unused bits.
		b = b << (8 - fromBits)
		if (b >> fromBits) > 0 {
			return nil, errors.New("Illegal bytes")
		}
		acc = ((acc << fromBits) | int(b)) & maxacc
		bits += int(fromBits)
		for {
			if bits < int(toBits) {
				break
			}
			bits -= int(toBits)
			results = append(results, (byte)((acc>>bits)&maxv))
		}
	}
	if pad && bits > 0 {
		results = append(results, (byte)((acc<<(int(toBits)-bits))&maxv))
	} else if bits >= int(fromBits) || (byte)((acc<<(int(toBits)-bits))&maxv) != 0 {
		return nil, errors.New("Illegal bytes")
	}

	return results, nil
}

func writeBech32Checksum(hrp string, data []byte) []byte {
	polymod := bech32Polymod(hrp, data, make([]byte, 6)) ^ 1
	checkSum := make([]byte, 6)
	for i := 0; i < 6; i++ {
		b := byte((polymod >> uint(5*(5-i))) & 31)

		// This can't fail, given we explicitly cap the previous b byte by the
		// first 31 bits.
		//c := charset[b]
		checkSum[i] = b
	}
	return checkSum
}

func bech32Polymod(hrp string, values, checksum []byte) int {
	chk := 1

	// Account for the high bits of the HRP in the checksum.
	for i := 0; i < len(hrp); i++ {
		b := chk >> 25
		hiBits := int(hrp[i]) >> 5
		chk = (chk&0x1ffffff)<<5 ^ hiBits
		for i := 0; i < 5; i++ {
			if (b>>uint(i))&1 == 1 {
				chk ^= gen[i]
			}
		}
	}

	// Account for the separator (0) between high and low bits of the HRP.
	// x^0 == x, so we eliminate the redundant xor used in the other rounds.
	b := chk >> 25
	chk = (chk & 0x1ffffff) << 5
	for i := 0; i < 5; i++ {
		if (b>>uint(i))&1 == 1 {
			chk ^= gen[i]
		}
	}

	// Account for the low bits of the HRP.
	for i := 0; i < len(hrp); i++ {
		b := chk >> 25
		loBits := int(hrp[i]) & 31
		chk = (chk&0x1ffffff)<<5 ^ loBits
		for i := 0; i < 5; i++ {
			if (b>>uint(i))&1 == 1 {
				chk ^= gen[i]
			}
		}
	}

	// Account for the values.
	for _, v := range values {
		b := chk >> 25
		chk = (chk&0x1ffffff)<<5 ^ int(v)
		for i := 0; i < 5; i++ {
			if (b>>uint(i))&1 == 1 {
				chk ^= gen[i]
			}
		}
	}

	if checksum == nil {
		// A nil checksum is used during encoding, so assume all bytes are zero.
		// x^0 == x, so we eliminate the redundant xor used in the other rounds.
		for v := 0; v < 6; v++ {
			b := chk >> 25
			chk = (chk & 0x1ffffff) << 5
			for i := 0; i < 5; i++ {
				if (b>>uint(i))&1 == 1 {
					chk ^= gen[i]
				}
			}
		}
	} else {
		// Checksum is provided during decoding, so use it.
		for _, v := range checksum {
			b := chk >> 25
			chk = (chk&0x1ffffff)<<5 ^ int(v)
			for i := 0; i < 5; i++ {
				if (b>>uint(i))&1 == 1 {
					chk ^= gen[i]
				}
			}
		}
	}

	return chk
}

var gen = []int{0x3b6a57b2, 0x26508e6d, 0x1ea119fa, 0x3d4233dd, 0x2a1462b3}

const charset = "qpzry9x8gf2tvdw0s3jn54khce6mua7l"

func readAddressType(addressBytes []byte) (*AddressType, error) {
	header := addressBytes[0]

	b := header & byte(0xF0)
	var addressType AddressType

	switch b >> 4 {
	case 0b0000: //base
	case 0b0001:
	case 0b0010:
	case 0b0011:
		addressType = Address_Type_Base
		break

	case 0b0100: //pointer
	case 0b0101:
		addressType = Address_Type_Ptr
		break
	case 0b0110: //enterprise
	case 0b0111:
		addressType = Address_Type_Enterprise
		break
	case 0b1110: //reward
	case 0b1111:
		addressType = Address_Type_Reward
		break
	case 0b1000: //byron
		addressType = Address_Type_Byron
		break
	default:
		return nil, errors.New("")
	}

	return &addressType, nil

}
func readNetwork(addressBytes []byte) (*Network, error) {
	header := addressBytes[0]
	var network Network
	switch header & 0x0f {
	case 0x00:
		network = TestNet
		break
	case 0x01:
		network = MainNet
		break
	default:
		return nil, errors.New("Unknown network type")
	}
	return &network, nil

}

func getAddressHeader(headerKind byte, network Network) byte {
	return byte(uint32(headerKind) | network.NetworkId&uint32(0xF))

}

func getPrefixHeader(addressType AddressType) string {
	if addressType == Address_Type_Byron {
		return "byr1"
	}
	if addressType == Address_Type_Reward {
		return "stake"
	} else {
		return "addr"
	}
}

func getPrefixTail(network Network) string {
	if network.NetworkId == MainNet.NetworkId {
		return ""
	} else {
		return "_test"
	}
}

func GetAddressBytes(address string) ([]byte, error) {
	if strings.HasPrefix(address, "stake") || strings.HasPrefix(address, "addr") {
		_, bytes, err := bech32.DecodeToBase256(address)
		if err != nil {
			return nil, err
		}
		return bytes, nil
	} else {
		decode := base58.Decode(address)
		return decode, nil
	}

}

func ByteToAddress(addressbyte []byte) (string, error) {
	addressType, err := readAddressType(addressbyte)
	if err != nil {
		return "", nil
	}
	if *addressType == Address_Type_Byron {
		return base58.Encode(addressbyte), nil
	} else {

		network, err := readNetwork(addressbyte)
		if err != nil {
			return "", nil
		}
		s := getPrefixHeader(*addressType) + getPrefixTail(*network)
		return encodeSegWitAddress(s, []byte("1")[0], addressbyte)
	}
}
