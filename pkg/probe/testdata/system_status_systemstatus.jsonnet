# api/v2.0/system/status.systemstatus
{
	"results": {
		"cluster": "fortiweb01",
		"haMultiGroup": false,
		"cluster_members": [
			{
				"hostname": "fortiweb01",
				"dev_sn": "FWBVM00000000001",
				"role": "Primary"
			},
			{
				"hostname": "fortiweb02",
				"dev_sn": "FWBVM00000000002",
				"role": "Secondary"
			}
		],
		"serialNumber": "FWBVM00000000001",
		"operationMode": "Reverse Proxy",
		"haStatus": "Active-Active-High-Volume",
		"systemTime": "Mon Sep 28 12:58:32 2026\n",
		"firmwareVersion": "FortiWeb-Azure 7.6.7,build1111(GA.M),260204",
		"up_days": "171",
		"up_hrs": "4",
		"up_mins": "49",
		"firmware_partition": 2,
		"administrativeDomain": "Enabled",
		"threatanalytics": "Disabled",
		"advancedBotProtection": "Disabled",
		"advancedBotProtectionAccountStatus": "License Pending",
		"vmLicense": "valid",
		"registration": {
			"label": "admin@example.com",
			"url": "https://support.fortinet.com",
			"text": "[Login]"
		},
		"readonly": false,
		"bufferSizeMax": 102400,
		"fileUploadLimitMax": 102400
	}
}