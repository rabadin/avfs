// Code generated by "stringer -type Features -trimprefix Feat -bitmask -output avfs_features.go"; DO NOT EDIT.

package avfs

import "strconv"

func _() {
	// An "invalid array index" compiler error signifies that the constant values have changed.
	// Re-run the stringer command to generate them again.
	var x [1]struct{}
	_ = x[FeatChroot-1]
	_ = x[FeatChownUser-2]
	_ = x[FeatSystemDirs-4]
	_ = x[FeatHardlink-8]
	_ = x[FeatIdentityMgr-16]
	_ = x[FeatReadOnly-32]
	_ = x[FeatReadOnlyIdm-64]
	_ = x[FeatRealFS-128]
	_ = x[FeatSymlink-256]
}

const _Features_name = "ChrootChownUserSystemDirsHardlinkIdentityMgrReadOnlyReadOnlyIdmRealFSSymlink"

var _Features_map = map[Features]string{
	1:   _Features_name[0:6],
	2:   _Features_name[6:15],
	4:   _Features_name[15:25],
	8:   _Features_name[25:33],
	16:  _Features_name[33:44],
	32:  _Features_name[44:52],
	64:  _Features_name[52:63],
	128: _Features_name[63:69],
	256: _Features_name[69:76],
}

func (i Features) String() string {
	if i <= 0 {
		return "Features()"
	}
	sb := make([]byte, 0, len(_Features_name)/2)
	sb = append(sb, []byte("Features(")...)
	for mask := Features(1); mask > 0 && mask <= i; mask <<= 1 {
		val := i & mask
		if val == 0 {
			continue
		}
		str, ok := _Features_map[val]
		if !ok {
			str = "0x" + strconv.FormatUint(uint64(val), 16)
		}
		sb = append(sb, []byte(str)...)
		sb = append(sb, '|')
	}
	sb[len(sb)-1] = ')'
	return string(sb)
}
