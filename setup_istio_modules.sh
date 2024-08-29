#!/bin/bash

# Define the kernel modules required for Istio's iptables interception mode
modules=(
    br_netfilter
    iptable_mangle
    iptable_nat
    iptable_raw
    xt_REDIRECT
    xt_connmark
    xt_conntrack
    xt_mark
    xt_owner
    xt_tcpudp
    xt_multiport
    bridge
    ip_tables
    nf_conntrack
    nf_nat
    x_tables
)

# Define IPv6-specific modules (only add if IPv6 is supported)
ipv6_modules=(
    ip6table_mangle
    ip6table_nat
    ip6table_raw
    ip6_tables
    nf_conntrack_ipv6
    nf_nat_ipv6
)

# Check if the system supports IPv6 before attempting to load IPv6 modules
if [ -f /proc/sys/net/ipv6/conf/all/disable_ipv6 ]; then
    if [ "$(sysctl -n net.ipv6.conf.all.disable_ipv6)" = "0" ]; then
        modules+=("${ipv6_modules[@]}")
    fi
else
    echo "IPv6 is not supported on this system; skipping IPv6-specific modules."
fi

# Load a module and ensure it persists across reboots
load_module() {
    local module=$1
    if ! lsmod | grep -q "$module"; then
        echo "Loading module: $module"
        if modprobe $module; then
            echo "$module" >> /etc/modules-load.d/istio.conf
        else
            echo "Failed to load module: $module"
        fi
    else
        echo "Module already loaded: $module"
    fi
}

# Ensure the /etc/modules-load.d/istio.conf file exists and is empty
echo -n "" > /etc/modules-load.d/istio.conf

# Load each required kernel module
for module in "${modules[@]}"; do
    load_module $module
done

echo "Kernel modules setup for Istio completed."

