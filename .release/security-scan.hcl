# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

binary {
  secrets {
    all = true
  }
  go_modules   = true
  osv          = true
  oss_index    = false
  nvd          = false
}
