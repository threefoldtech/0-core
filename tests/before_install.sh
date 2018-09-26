export ubuntu_port=$((2500 + RANDOM % 1000))
export vm_ubuntu_name=$(head /dev/urandom | tr -dc A-Za-z0-9 | head -c 10 ; echo '')
export vm_zos_name=$(head /dev/urandom | tr -dc A-Za-z0-9 | head -c 10 ; echo '')
export bridge=$(head /dev/urandom | tr -dc A-Za-z0-9 | head -c 10 ; echo '')
