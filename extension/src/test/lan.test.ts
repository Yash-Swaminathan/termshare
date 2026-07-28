import { test } from "node:test";
import assert from "node:assert/strict";
import { detectHostLANIP, lanIPScore } from "../lan";

test("lanIPScore prefers 192.168 over 10 and 172", () => {
  assert.ok(lanIPScore("192.168.1.5") > lanIPScore("10.0.0.5"));
  assert.ok(lanIPScore("10.0.0.5") > lanIPScore("172.17.156.134"));
  assert.equal(lanIPScore("not-an-ip"), 0);
});

test("detectHostLANIP prefers Wi-Fi over VirtualBox and gateway .1", () => {
  const got = detectHostLANIP({
    "vEthernet (WSL)": [
      { address: "172.17.144.1", netmask: "255.255.240.0", family: "IPv4", mac: "", internal: false, cidr: null }
    ],
    "Ethernet 2": [
      { address: "192.168.56.1", netmask: "255.255.255.0", family: "IPv4", mac: "", internal: false, cidr: null }
    ],
    "Wi-Fi": [
      { address: "192.168.86.100", netmask: "255.255.255.0", family: "IPv4", mac: "", internal: false, cidr: null }
    ]
  });
  assert.equal(got, "192.168.86.100");
});

test("detectHostLANIP ignores link-local", () => {
  const got = detectHostLANIP({
    "Wi-Fi": [
      { address: "169.254.10.20", netmask: "255.255.0.0", family: "IPv4", mac: "", internal: false, cidr: null }
    ]
  });
  assert.equal(got, undefined);
});
