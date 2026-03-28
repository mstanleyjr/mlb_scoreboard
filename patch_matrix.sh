#!/bin/bash
# Patch the go-rpi-rgb-led-matrix library to add gpio_slowdown and row_address_type support
# Run this on the Pi: bash patch_matrix.sh

LIB=~/go/pkg/mod/github.com/tfk1410/go-rpi-rgb-led-matrix@v0.0.0-20210404121211-ed43f29cbccb/matrix.go

echo "Making library writable..."
sudo chmod -R u+w ~/go/pkg/mod/github.com/tfk1410/go-rpi-rgb-led-matrix@v0.0.0-20210404121211-ed43f29cbccb/

echo "File check:"
ls -la $LIB

echo ""
echo "Current relevant lines:"
grep -n "inverse_colors\|InverseColors\|GPIOSlowdown\|gpio_slowdown\|HardwareMapping\|led_matrix_create" $LIB

echo ""
echo "Patching..."

python3 - <<'PYEOF'
import os, sys

home = os.path.expanduser("~")
path = f"{home}/go/pkg/mod/github.com/tfk1410/go-rpi-rgb-led-matrix@v0.0.0-20210404121211-ed43f29cbccb/matrix.go"

with open(path, 'r') as f:
    content = f.read()

if 'GPIOSlowdown' in content:
    print("Already patched!")
    sys.exit(0)

failed = False

# Step 1: Add C functions after set_inverse_colors
marker = 'o->inverse_colors = inverse_colors != 0 ? 1 : 0;\n}'
if marker not in content:
    print("ERROR Step 1: Can't find marker. Lines with 'inverse':")
    for i, l in enumerate(content.split('\n')):
        if 'inverse' in l.lower(): print(f"  {i}: {repr(l)}")
    failed = True
else:
    content = content.replace(marker,
        marker + '\n\nvoid set_gpio_slowdown(struct RGBLedRuntimeOptions *o, int gpio_slowdown) {\n  o->gpio_slowdown = gpio_slowdown;\n}\n\nvoid set_row_address_type(struct RGBLedMatrixOptions *o, int row_address_type) {\n  o->row_address_type = row_address_type;\n}', 1)
    print("Step 1 OK")

# Step 2: Add struct fields - find HardwareMapping line and the closing brace
import re
m = re.search(r'([ \t]+HardwareMapping string\n)(})', content)
if not m:
    print("ERROR Step 2: Can't find HardwareMapping field")
    failed = True
else:
    indent = re.match(r'[ \t]+', m.group(1)).group()
    insert = f"{indent}// GPIOSlowdown slows GPIO speed to reduce flicker. Pi 4B = 3\n{indent}GPIOSlowdown int\n{indent}// RowAddressType: 0=default, 1=AB-addressed (64-row panels)\n{indent}RowAddressType int\n"
    content = content[:m.start(2)] + insert + content[m.start(2):]
    print("Step 2 OK")

# Step 3: Add RowAddressType call before return o in toC()
m = re.search(r'([ \t]+)(return o\n})', content)
if not m:
    print("ERROR Step 3: Can't find 'return o'")
    failed = True
else:
    indent = m.group(1)
    insert = f"\n{indent}if c.RowAddressType != 0 {{\n{indent}\tC.set_row_address_type(o, C.int(c.RowAddressType))\n{indent}}}\n\n"
    content = content[:m.start(2)] + insert + content[m.start(2):]
    print("Step 3 OK")

# Step 4: Add RuntimeOptions to NewRGBLedMatrix
m = re.search(r'([ \t]+)(m := C\.led_matrix_create_from_options\(config\.toC\(\), nil, nil\))', content)
if not m:
    print("ERROR Step 4: Can't find led_matrix_create_from_options")
    failed = True
else:
    indent = m.group(1)
    new_lines = f"{indent}runtimeOpts := &C.struct_RGBLedRuntimeOptions{{}}\n{indent}runtimeOpts.gpio_slowdown = C.int(config.GPIOSlowdown)\n{indent}m := C.led_matrix_create_from_options(config.toC(), runtimeOpts, nil)"
    content = content[:m.start()] + new_lines + content[m.end():]
    print("Step 4 OK")

if failed:
    print("\nPatch FAILED - see errors above")
    sys.exit(1)

with open(path, 'w') as f:
    f.write(content)

print("\nVerifying...")
with open(path, 'r') as f:
    v = f.read()
all_ok = True
for check in ['GPIOSlowdown', 'RowAddressType', 'set_gpio_slowdown', 'set_row_address_type', 'runtimeOpts']:
    ok = check in v
    print(f"  {'✓' if ok else '✗'} {check}")
    if not ok: all_ok = False

if all_ok:
    print("\nPatch applied successfully!")
else:
    print("\nPatch INCOMPLETE - some checks failed")
    sys.exit(1)
PYEOF

echo "Done! Now update your main.go to use GPIOSlowdown = 3 and reboot."

