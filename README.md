Airtame-CLI
===========
The title is somewhat misleading. The idea is that it lets you do stuff you would usually do at https://airtame.cloud in the CLI, but right now all you can do with this is
 - List info about all devices by group
 - List info about all devices, but flattened
 - Reboot a specific Airtame
 - Reboot all the Airtames in your account

This exists primarily for that last one. We suspect that it might be necessary to reboot these devices regularly, so we wanted to set up something to do that every night. I only intend to add features as I need them, but PRs welcome.

Usage
=====
```sh
airtame-cli --email jon@example.com --password mypassword devices
airtame-cli --email jon@example.com --password mypassword flatdevices
airtame-cli --email jon@example.com --password mypassword reboot 259223
airtame-cli --email jon@example.com --password mypassword rebootall
```

Example Output (devices by group)
=================================
```json
[
	{
		"groupId": 40229,
		"groupName": "Admin-Office",
		"organizationId": 8444,
		"devices": [
			{
				"platform": "DG2",
				"version": "v4.4.3",
				"state": "DMGR_IDLE",
				"screenshotEnabled": false,
				"settingsState": "synced",
				"id": 259223,
				"ap24Enabled": false,
				"ap24Channel": 6,
				"ap52Enabled": false,
				"ap52Channel": 40,
				"backgroundType": "Custom image",
				"deviceName": "Conf-233",
				"lastSeen": 7,
				"lastConnected": 1635200415,
				"isOnline": true,
				"networkState": {
					"online": true,
					"interfaces": [
						{
							"frequency": 5785,
							"ip": "10.7.18.85",
							"mac": "38:4b:76:89:22:b6",
							"mode": "client",
							"name": "wlan0",
							"signal_strength": -58,
							"ssid": "marconi",
							"status": "connected",
							"type": "wifi"
						},
						{
							"frequency": 0,
							"ip": "169.254.13.37",
							"mac": "38:4b:76:89:22:b6",
							"mode": "",
							"name": "usb0",
							"signal_strength": 0,
							"ssid": "",
							"status": "connected",
							"type": "usb"
						},
						{
							"frequency": 0,
							"ip": "",
							"mac": "52:96:a5:89:22:42",
							"mode": "",
							"name": "br0",
							"signal_strength": 0,
							"ssid": "",
							"status": "disconnected",
							"type": "bridge"
						}
					]
				},
				"updateAvailable": false,
				"updateChannel": "ga",
				"updateProgress": 0,
				"homescreenOrientation": "landscape"
			}
		]
	}
]

```
