// generated by go run gen.go; DO NOT EDIT

package publicsuffix

const version = "subset of publicsuffix.org's effective_tld_names.dat, hg revision 05b11a8d1ace (2012-11-09)"

const (
	nodesBitsChildren   = 9
	nodesBitsICANN      = 1
	nodesBitsTextOffset = 15
	nodesBitsTextLength = 6

	childrenBitsWildcard = 1
	childrenBitsNodeType = 2
	childrenBitsHi       = 14
	childrenBitsLo       = 14
)

const (
	nodeTypeNormal     = 0
	nodeTypeException  = 1
	nodeTypeParentOnly = 2
)

// numTLD is the number of top level domains.
const numTLD = 13

// Text is the combined text of all labels.
const text = "clubacyfukuchiyamashinacionakagyomanpostatecouncilgovgvhomediaph" +
	"onelin-addretinagaokakyotambainetip6irischigashiyamaizurujitawar" +
	"akpetroleumiljetjoyomanteljpblogspotk12kizurideduccitykumiyamaky" +
	"otangobiernoelectronicorgamecongresodelalengua3kyotoyamazakitami" +
	"namiyamashiromiyazurnantanational-library-scotlandyndnsakyotanab" +
	"ebizwmukobenlseikameokamodpromocionawrastelecomanmobileusiemenso" +
	"ngfestxn--czrw28british-libraryawatarparliamentwazukayabe164xn--" +
	"p1aidvxn--uc0atvxn--zf0ao64a"

// nodes is the list of nodes. Each node is represented as a uint32, which
// encodes the node's children, wildcard bit and node type (as an index into
// the children array), ICANN bit and text.
//
// In the //-comment after each node's data, the nodes indexes of the children
// are formatted as (n0x1234-n0x1256), with * denoting the wildcard bit. The
// nodeType is printed as + for normal, ! for exception, and o for parent-only
// nodes that have children but don't match a domain label in their own right.
// An I denotes an ICANN domain.
//
// The layout within the uint32, from MSB to LSB, is:
//	[ 1 bits] unused
//	[ 9 bits] children index
//	[ 1 bits] ICANN bit
//	[15 bits] text index
//	[ 6 bits] text length
var nodes = [...]uint32{
	0x01a01442, // n0x0000 c0x0006 (n0x000d-n0x0013)  + I ao
	0x01e01f82, // n0x0001 c0x0007 (n0x0013-n0x001d)* o I ar
	0x026068c4, // n0x0002 c0x0009 (n0x001e-n0x0024)  o I arpa
	0x02a05b03, // n0x0003 c0x000a (n0x0024-n0x0025)  o I com
	0x01600142, // n0x0004 c0x0005 (---------------)* o I cy
	0x02e02682, // n0x0005 c0x000b (n0x0025-n0x0028)  + I jp
	0x03a00802, // n0x0006 c0x000e (n0x0048-n0x0052)* o I om
	0x03e03583, // n0x0007 c0x000f (n0x0052-n0x0053)  o I org
	0x04606b82, // n0x0008 c0x0011 (n0x0055-n0x0063)  + I tw
	0x04a00202, // n0x0009 c0x0012 (n0x0063-n0x006e)* o I uk
	0x05205dc2, // n0x000a c0x0014 (n0x006f-n0x0070)  o I us
	0x00206f08, // n0x000b c0x0000 (---------------)  + I xn--p1ai
	0x016050c2, // n0x000c c0x0005 (---------------)* o I zw
	0x00200ac2, // n0x000d c0x0000 (---------------)  + I co
	0x00200e82, // n0x000e c0x0000 (---------------)  + I ed
	0x00200d42, // n0x000f c0x0000 (---------------)  + I gv
	0x00201e82, // n0x0010 c0x0000 (---------------)  + I it
	0x00202782, // n0x0011 c0x0000 (---------------)  + I og
	0x002026c2, // n0x0012 c0x0000 (---------------)  + I pb
	0x02205b03, // n0x0013 c0x0008 (n0x001d-n0x001e)  o I com
	0x00603713, // n0x0014 c0x0001 (---------------)  ! I congresodelalengua3
	0x00602b84, // n0x0015 c0x0001 (---------------)  ! I educ
	0x00603113, // n0x0016 c0x0001 (---------------)  ! I gobiernoelectronico
	0x00603685, // n0x0017 c0x0001 (---------------)  ! I mecon
	0x00600546, // n0x0018 c0x0001 (---------------)  ! I nacion
	0x006034c3, // n0x0019 c0x0001 (---------------)  ! I nic
	0x00605689, // n0x001a c0x0001 (---------------)  ! I promocion
	0x00601286, // n0x001b c0x0001 (---------------)  ! I retina
	0x00600083, // n0x001c c0x0001 (---------------)  ! I uba
	0x00002708, // n0x001d c0x0000 (---------------)  +   blogspot
	0x00206e04, // n0x001e c0x0000 (---------------)  + I e164
	0x00201107, // n0x001f c0x0000 (---------------)  + I in-addr
	0x00201843, // n0x0020 c0x0000 (---------------)  + I ip6
	0x00201904, // n0x0021 c0x0000 (---------------)  + I iris
	0x00202a83, // n0x0022 c0x0000 (---------------)  + I uri
	0x00204483, // n0x0023 c0x0000 (---------------)  + I urn
	0x00000202, // n0x0024 c0x0000 (---------------)  +   uk
	0x00200102, // n0x0025 c0x0000 (---------------)  + I ac
	0x032051c4, // n0x0026 c0x000c (n0x0028-n0x0029)* o I kobe
	0x03603bc5, // n0x0027 c0x000d (n0x0029-n0x0048)  + I kyoto
	0x00602c84, // n0x0028 c0x0001 (---------------)  ! I city
	0x00206d05, // n0x0029 c0x0000 (---------------)  + I ayabe
	0x002001cb, // n0x002a c0x0000 (---------------)  + I fukuchiyama
	0x00201a4b, // n0x002b c0x0000 (---------------)  + I higashiyama
	0x00202b03, // n0x002c c0x0000 (---------------)  + I ide
	0x00201743, // n0x002d c0x0000 (---------------)  + I ine
	0x00202404, // n0x002e c0x0000 (---------------)  + I joyo
	0x00205407, // n0x002f c0x0000 (---------------)  + I kameoka
	0x00205544, // n0x0030 c0x0000 (---------------)  + I kamo
	0x00203e84, // n0x0031 c0x0000 (---------------)  + I kita
	0x002029c4, // n0x0032 c0x0000 (---------------)  + I kizu
	0x00202d88, // n0x0033 c0x0000 (---------------)  + I kumiyama
	0x00201548, // n0x0034 c0x0000 (---------------)  + I kyotamba
	0x00204e09, // n0x0035 c0x0000 (---------------)  + I kyotanabe
	0x00202f88, // n0x0036 c0x0000 (---------------)  + I kyotango
	0x00201c87, // n0x0037 c0x0000 (---------------)  + I maizuru
	0x00203f86, // n0x0038 c0x0000 (---------------)  + I minami
	0x00203f8f, // n0x0039 c0x0000 (---------------)  + I minamiyamashiro
	0x00204346, // n0x003a c0x0000 (---------------)  + I miyazu
	0x00205144, // n0x003b c0x0000 (---------------)  + I muko
	0x0020138a, // n0x003c c0x0000 (---------------)  + I nagaokakyo
	0x00200687, // n0x003d c0x0000 (---------------)  + I nakagyo
	0x00204506, // n0x003e c0x0000 (---------------)  + I nantan
	0x00203cc9, // n0x003f c0x0000 (---------------)  + I oyamazaki
	0x00204d85, // n0x0040 c0x0000 (---------------)  + I sakyo
	0x00205345, // n0x0041 c0x0000 (---------------)  + I seika
	0x00204ec6, // n0x0042 c0x0000 (---------------)  + I tanabe
	0x00201e03, // n0x0043 c0x0000 (---------------)  + I uji
	0x00201e09, // n0x0044 c0x0000 (---------------)  + I ujitawara
	0x00206bc6, // n0x0045 c0x0000 (---------------)  + I wazuka
	0x00200389, // n0x0046 c0x0000 (---------------)  + I yamashina
	0x00206786, // n0x0047 c0x0000 (---------------)  + I yawata
	0x00600e4a, // n0x0048 c0x0001 (---------------)  ! I mediaphone
	0x00605886, // n0x0049 c0x0001 (---------------)  ! I nawras
	0x0060588d, // n0x004a c0x0001 (---------------)  ! I nawrastelecom
	0x00605b4a, // n0x004b c0x0001 (---------------)  ! I omanmobile
	0x00600808, // n0x004c c0x0001 (---------------)  ! I omanpost
	0x006024c7, // n0x004d c0x0001 (---------------)  ! I omantel
	0x00601fcc, // n0x004e c0x0001 (---------------)  ! I rakpetroleum
	0x00605e07, // n0x004f c0x0001 (---------------)  ! I siemens
	0x00605f88, // n0x0050 c0x0001 (---------------)  ! I songfest
	0x0060098c, // n0x0051 c0x0001 (---------------)  ! I statecouncil
	0x04004c46, // n0x0052 c0x0010 (n0x0053-n0x0055)  +   dyndns
	0x00000c82, // n0x0053 c0x0000 (---------------)  +   go
	0x00000dc4, // n0x0054 c0x0000 (---------------)  +   home
	0x00002708, // n0x0055 c0x0000 (---------------)  +   blogspot
	0x00200004, // n0x0056 c0x0000 (---------------)  + I club
	0x00205b03, // n0x0057 c0x0000 (---------------)  + I com
	0x00205004, // n0x0058 c0x0000 (---------------)  + I ebiz
	0x00202b83, // n0x0059 c0x0000 (---------------)  + I edu
	0x00203604, // n0x005a c0x0000 (---------------)  + I game
	0x00200c83, // n0x005b c0x0000 (---------------)  + I gov
	0x002070c3, // n0x005c c0x0000 (---------------)  + I idv
	0x00202283, // n0x005d c0x0000 (---------------)  + I mil
	0x00201783, // n0x005e c0x0000 (---------------)  + I net
	0x00203583, // n0x005f c0x0000 (---------------)  + I org
	0x0020618b, // n0x0060 c0x0000 (---------------)  + I xn--czrw28b
	0x0020718a, // n0x0061 c0x0000 (---------------)  + I xn--uc0atv
	0x0020740c, // n0x0062 c0x0000 (---------------)  + I xn--zf0ao64a
	0x00602702, // n0x0063 c0x0001 (---------------)  ! I bl
	0x0060640f, // n0x0064 c0x0001 (---------------)  ! I british-library
	0x04e00ac2, // n0x0065 c0x0013 (n0x006e-n0x006f)  o I co
	0x00602343, // n0x0066 c0x0001 (---------------)  ! I jet
	0x006055c3, // n0x0067 c0x0001 (---------------)  ! I mod
	0x00604659, // n0x0068 c0x0001 (---------------)  ! I national-library-scotland
	0x00601043, // n0x0069 c0x0001 (---------------)  ! I nel
	0x006034c3, // n0x006a c0x0001 (---------------)  ! I nic
	0x006052c3, // n0x006b c0x0001 (---------------)  ! I nls
	0x0060694a, // n0x006c c0x0001 (---------------)  ! I parliament
	0x016019c3, // n0x006d c0x0005 (---------------)* o I sch
	0x00002708, // n0x006e c0x0000 (---------------)  +   blogspot
	0x056006c2, // n0x006f c0x0015 (n0x0070-n0x0073)  + I ak
	0x00202c42, // n0x0070 c0x0000 (---------------)  + I cc
	0x00202903, // n0x0071 c0x0000 (---------------)  + I k12
	0x00204883, // n0x0072 c0x0000 (---------------)  + I lib
}

// children is the list of nodes' children, the parent's wildcard bit and the
// parent's node type. If a node has no children then their children index
// will be in the range [0, 6), depending on the wildcard bit and node type.
//
// The layout within the uint32, from MSB to LSB, is:
//	[ 1 bits] unused
//	[ 1 bits] wildcard bit
//	[ 2 bits] node type
//	[14 bits] high nodes index (exclusive) of children
//	[14 bits] low nodes index (inclusive) of children
var children = [...]uint32{
	0x00000000, // c0x0000 (---------------)  +
	0x10000000, // c0x0001 (---------------)  !
	0x20000000, // c0x0002 (---------------)  o
	0x40000000, // c0x0003 (---------------)* +
	0x50000000, // c0x0004 (---------------)* !
	0x60000000, // c0x0005 (---------------)* o
	0x0004c00d, // c0x0006 (n0x000d-n0x0013)  +
	0x60074013, // c0x0007 (n0x0013-n0x001d)* o
	0x2007801d, // c0x0008 (n0x001d-n0x001e)  o
	0x2009001e, // c0x0009 (n0x001e-n0x0024)  o
	0x20094024, // c0x000a (n0x0024-n0x0025)  o
	0x000a0025, // c0x000b (n0x0025-n0x0028)  +
	0x600a4028, // c0x000c (n0x0028-n0x0029)* o
	0x00120029, // c0x000d (n0x0029-n0x0048)  +
	0x60148048, // c0x000e (n0x0048-n0x0052)* o
	0x2014c052, // c0x000f (n0x0052-n0x0053)  o
	0x00154053, // c0x0010 (n0x0053-n0x0055)  +
	0x0018c055, // c0x0011 (n0x0055-n0x0063)  +
	0x601b8063, // c0x0012 (n0x0063-n0x006e)* o
	0x201bc06e, // c0x0013 (n0x006e-n0x006f)  o
	0x201c006f, // c0x0014 (n0x006f-n0x0070)  o
	0x001cc070, // c0x0015 (n0x0070-n0x0073)  +
}
