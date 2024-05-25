package globals

type Code struct {
	Checksum string `yaml:"checksum"`
	Source   string `yaml:"source"`
}

var Registry = map[string]Code{
	"calc": {
		Checksum: "DD9761B3E49D34BE18076CFA2BE7AC4F8B3AF65EB13CB8B24E38F9388130C2D2",
		Source:   "kaiyo-1://157",
	},
	"kujira_bow_lsd": {
		Checksum: "715934C48A41A7AEC050A6DEC9A1C957ABF96FFC84F7B8AD3260309749E9A7E9",
		Source:   "kaiyo-1://167",
	},
	"kujira_bow_stable": {
		Checksum: "F3B812304F6AECEBD4CE388DA0F5161FD5BE1589C9008C5036258EC89ABCC502",
		Source:   "kaiyo-1://166",
	},
	"kujira_bow_xyk": {
		Checksum: "418CF9A2E005B6B44352DAFE1E8C5998F834F3BACF47497E95E69A3EB2DFAA22",
		Source:   "kaiyo-1://126",
	},
	"kujira_bow_margin": {
		Checksum: "5EEEAEC66D81449EEE7A687C47CBE53AD1571AA9F235CD019BF354F65F3C4610",
		Source:   "kaiyo-1://188",
	},
	"kujira_bow_staking": {
		Checksum: "B081C828D3FC8FA658CA9906423BB90744336CCBA5363881A1CDB7449F716AC6",
		Source:   "kaiyo-1://244",
	},
	"kujira_fin": {
		Checksum: "8A6FA03E62DA9CB75F1CB9A4EFEA6AAFA920AD5FCA40A7B335560727BD42C198",
		Source:   "kaiyo-1://243",
	},
	"kujira_ghost_market": {
		Checksum: "F9D6AEC8CD94935C6EB260FF7E63F372E4C959FC849B92E5CAF21F3F4547DC6C",
		Source:   "kaiyo-1://291",
	},
	"kujira_ghost_vault": {
		Checksum: "47CF68C95E260931A99440361285A77978180131BE8AE8F1DA2A5473BF76D4BB",
		Source:   "kaiyo-1://316",
	},
	"kujira_orca": {
		Checksum: "64ECE3AB0E8CD640023FF81E6C7F5CF1C0C19D4CAFA419BD241976B6A714B2A7",
		Source:   "kaiyo-1://234",
	},
	"kujira_pilot": {
		Checksum: "A178E33C489142FF5DE8A89261BB3CA86EEFC2FB382DCD57CBB6FB1531D55F46",
		Source:   "kaiyo-1://95",
	},
	"kujira_stable_margin_swap": {
		Checksum: "2395063CC30A098DB87A7F275C7A8EBFF889E9D40DB18E09CED6921370351DC4",
		Source:   "kaiyo-1://87",
	},
	"kujira_stable_market": {
		Checksum: "E9EC73285A3D9E1CED23924D3BA6EA9267B039DA596686EA7AEB1E474415DA13",
		Source:   "kaiyo-1://73",
	},
	"kujira_stable_mint": {
		Checksum: "98CC2EDAA8A5D1AD8CD15EADB6CA8A52267F29142177F105E3C65D0ECFF50C5F",
		Source:   "kaiyo-1://11",
	},
}
