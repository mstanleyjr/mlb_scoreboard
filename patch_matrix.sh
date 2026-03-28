#!/bin/bash
# Patch the go-rpi-rgb-led-matrix library to add gpio_slowdown and row_address_type support
# Run this on the Pi: bash patch_matrix.sh

LIB=~/go/pkg/mod/github.com/tfk1410/go-rpi-rgb-led-matrix@v0.0.0-20210404121211-ed43f29cbccb/matrix.go

echo "Making library writable..."
sudo chmod -R u+w ~/go/pkg/mod/github.com/tfk1410/go-rpi-rgb-led-matrix@v0.0.0-20210404121211-ed43f29cbccb/

echo "Patching $LIB..."
echo "File exists check:"
ls -la $LIB

echo ""
echo "Current content around InverseColors:"
grep -n "inverse_colors\|InverseColors\|GPIOSlowdown\|gpio_slowdown" $LIB

python3 - <<'PYEOF'
import re

path = "/root/go/pkg/mod/github.com/tfk1410/go-rpi-rgb-led-matrix@v0.0.0-20210404121211-ed43f29cbccb/matrix.go"

# Try both root and home paths
import os
home = os.path.expanduser("~")
path = f"{home}/go/pkg/mod/github.com/tfk1410/go-rpi-rgb-led-matrix@v0.0.0-20210404121211-ed43f29cbccb/matrix.go"

with open(path, 'r') as f:
    content = f.read()

# Check if already patched
if 'set_gpio_slowdown' in content:
    print("Already patched!")
    exit(0)

# Add C functions before the closing */ of the CGO preamble
new_c_funcs = """
void set_gpio_slowdown(struct RGBLedRuntimeOptions *o, int gpio_slowdown) {
  o->gpio_slowdown = gpio_slowdown;
}

void set_row_address_type(struct RGBLedMatrixOptions *o, int row_address_type) {
  o->row_address_type = row_address_type;
}
"""

# Insert before the closing */ of the C block
content = content.replace(
    'void set_inverse_colors(struct RGBLedMatrixOptions *o, int inverse_colors) {\n  o->inverse_colors = inverse_colors != 0 ? 1 : 0;\n}',
    'void set_inverse_colors(struct RGBLedMatrixOptions *o, int inverse_colors) {\n  o->inverse_colors = inverse_colors != 0 ? 1 : 0;\n}' + new_c_funcs
)

# Add fields to HardwareConfig struct
content = content.replace(
    '        // Name of GPIO mapping used\n        HardwareMapping string\n}',
    '        // Name of GPIO mapping used\n        HardwareMapping string\n        // GPIOSlowdown slows down GPIO to reduce flicker. Use 1 for Pi 1/2, 2 for Pi 3, 3-4 for Pi 4\n        GPIOSlowdown int\n        // RowAddressType 0=default, 1=AB-addressed panels (required for 64-row panels)\n        RowAddressType int\n}'
)

# Add to toC() function - find the return statement and insert before it
content = content.replace(
    '        if c.InverseColors == true {\n                C.set_inverse_colors(o, C.int(1))\n        } else {\n                C.set_inverse_colors(o, C.int(0))\n        }\n\n        return o',
    '        if c.InverseColors == true {\n                C.set_inverse_colors(o, C.int(1))\n        } else {\n                C.set_inverse_colors(o, C.int(0))\n        }\n\n        if c.RowAddressType != 0 {\n                C.set_row_address_type(o, C.int(c.RowAddressType))\n        }\n\n        return o'
)

# Add RuntimeOptions handling in NewRGBLedMatrix
# Find the matrix creation call and add runtime options
content = content.replace(
    '        m := C.led_matrix_create_from_options(config.toC(), nil, nil)',
    '        runtimeOpts := &C.struct_RGBLedRuntimeOptions{}\n        if config.GPIOSlowdown > 0 {\n                C.set_gpio_slowdown(runtimeOpts, C.int(config.GPIOSlowdown))\n        }\n        m := C.led_matrix_create_from_options(config.toC(), runtimeOpts, nil)'
)

with open(path, 'w') as f:
    f.write(content)

print("Patch applied successfully!")
PYEOF

echo "Done! Now update your main.go to use GPIOSlowdown = 3 and reboot."

