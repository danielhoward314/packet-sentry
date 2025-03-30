#!/bin/bash
# Prompt user for install key if not provided
if [ ! -f "/opt/packet-sentry" ]; then
    echo "Please enter the install key:"

    # Save current value of shell history file to temp variable
    OLD_HISTFILE="${HISTFILE}"
    
    # Unset HISTFILE temporarily
    HISTFILE=
    
    # Execute the commands that should be excluded from history
    read -s -r install_key
    cat <<-EOF > "/opt/packet-sentry"
    {
        "installKey": "$install_key"
    }
EOF

    # Restore HISTFILE to its original value
    HISTFILE="${OLD_HISTFILE}"

    chmod 600 "/opt/packet-sentry"
    echo "Install key has been saved."
fi