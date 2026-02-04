// Copyright 2026 Palantir Technologies, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package testplugin

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/pkg/errors"
)

// Partition represents a partition configuration for splitting packages.
type Partition struct {
	Index int // 0-indexed partition number
	Total int // total number of partitions
}

// ParsePartition parses a partition string in the format "X,N" where X is the
// 0-indexed partition number and N is the total number of partitions.
// Returns nil if the input is empty (no partitioning).
func ParsePartition(s string) (*Partition, error) {
	if s == "" {
		return nil, nil
	}
	parts := strings.Split(s, ",")
	if len(parts) != 2 {
		return nil, errors.Errorf("invalid partition format %q: expected format X,N (e.g., 0,4)", s)
	}
	index, err := strconv.Atoi(strings.TrimSpace(parts[0]))
	if err != nil {
		return nil, errors.Wrapf(err, "invalid partition index %q", parts[0])
	}
	total, err := strconv.Atoi(strings.TrimSpace(parts[1]))
	if err != nil {
		return nil, errors.Wrapf(err, "invalid partition total %q", parts[1])
	}
	if total < 1 {
		return nil, errors.Errorf("partition total must be at least 1, got %d", total)
	}
	if index < 0 || index >= total {
		return nil, errors.Errorf("partition index must be between 0 and %d, got %d", total-1, index)
	}
	return &Partition{Index: index, Total: total}, nil
}

// Apply applies the partition to the given slice of packages, returning only
// the packages that belong to this partition. Packages are sorted before
// partitioning to ensure deterministic results.
// Returns the original slice if partition is nil.
func (p *Partition) Apply(pkgs []string) []string {
	if p == nil || len(pkgs) == 0 {
		return pkgs
	}
	// Sort for deterministic partitioning
	sorted := make([]string, len(pkgs))
	copy(sorted, pkgs)
	sort.Strings(sorted)

	// Calculate partition boundaries
	totalPkgs := len(sorted)
	baseSize := totalPkgs / p.Total
	remainder := totalPkgs % p.Total

	// Distribute packages: first 'remainder' partitions get baseSize+1 packages,
	// remaining partitions get baseSize packages
	start := 0
	for i := 0; i < p.Index; i++ {
		if i < remainder {
			start += baseSize + 1
		} else {
			start += baseSize
		}
	}

	size := baseSize
	if p.Index < remainder {
		size = baseSize + 1
	}

	end := start + size
	if start >= totalPkgs {
		return nil
	}
	if end > totalPkgs {
		end = totalPkgs
	}
	return sorted[start:end]
}

// String returns a human-readable string representation of the partition.
func (p *Partition) String() string {
	if p == nil {
		return "no partition"
	}
	return fmt.Sprintf("partition %d of %d", p.Index, p.Total)
}
