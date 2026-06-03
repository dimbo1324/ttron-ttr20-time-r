package mapping

import ft12v1 "github.com/dimbo1324/ttron-ttr20-time-r/internal/api/grpc/ft12/v1"

func ChecksumMode(mode string) ft12v1.ChecksumMode {
	switch mode {
	case "sum":
		return ft12v1.ChecksumMode_CHECKSUM_MODE_SUM
	case "crc16":
		return ft12v1.ChecksumMode_CHECKSUM_MODE_CRC16
	default:
		return ft12v1.ChecksumMode_CHECKSUM_MODE_UNSPECIFIED
	}
}
