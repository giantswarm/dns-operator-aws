package route53

var (
	// see: https://docs.aws.amazon.com/general/latest/gr/elb.html
	canonicalHostedZones = map[string]string{
		// Application Load Balancers and Classic Load Balancers
		"us-east-2":      "Z3AADJGX6KTTL2",
		"us-east-1":      "Z35SXDOTRQ7X7K",
		"us-west-1":      "Z368ELLRRE2KJ0",
		"us-west-2":      "Z1H1FL5HABSF5",
		"ca-central-1":   "ZQSVJUPU6J1EY",
		"ap-east-1":      "Z3DQVH9N71FHZ0",
		"ap-south-1":     "ZP97RAFLXTNZK",
		"ap-northeast-2": "ZWKZPGTI48KDX",
		"ap-northeast-3": "Z5LXEXXYW11ES",
		"ap-southeast-1": "Z1LMS91P8CMLE5",
		"ap-southeast-2": "Z1GM3OXH4ZPM65",
		"ap-northeast-1": "Z14GRHDCWA56QT",
		"eu-central-1":   "Z215JYRZR1TBD5",
		"eu-west-1":      "Z32O12XQLNTSW2",
		"eu-west-2":      "ZHURV8PSTC4K8",
		"eu-west-3":      "Z3Q77PNBQS71R4",
		"eu-north-1":     "Z23TAZ6LKFMNIO",
		"eu-south-1":     "Z3ULH7SSC9OV64",
		"sa-east-1":      "Z2P70J7HTTTPLU",
		"cn-north-1":     "Z1GDH35T77C1KE",
		"cn-northwest-1": "ZM7IZAIOVVDZF",
		"us-gov-west-1":  "Z33AYJ8TM3BH4J",
		"us-gov-east-1":  "Z166TLBEWOO7G0",
		"me-south-1":     "ZS929ML54UICD",
		"af-south-1":     "Z268VQBMOI5EKX",
	}
)
