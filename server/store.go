// Copyright 2016 PingCAP, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// See the License for the specific language governing permissions and
// limitations under the License.

package server

import (
	"time"

	"github.com/gogo/protobuf/proto"
	"github.com/pingcap/kvproto/pkg/metapb"
	"github.com/pingcap/kvproto/pkg/pdpb"
)

// ResourceKind distinguishes different kinds of resources.
type ResourceKind int

const (
	leaderKind ResourceKind = iota + 1
	regionKind
)

// storeInfo contains information about a store.
// TODO: Export this to API directly.
type storeInfo struct {
	*metapb.Store
	stats *StoreStatus
}

func newStoreInfo(store *metapb.Store) *storeInfo {
	return &storeInfo{
		Store: store,
		stats: newStoreStatus(),
	}
}

func (s *storeInfo) clone() *storeInfo {
	return &storeInfo{
		Store: proto.Clone(s.Store).(*metapb.Store),
		stats: s.stats.clone(),
	}
}

func (s *storeInfo) block() {
	s.stats.blocked = true
}

func (s *storeInfo) unblock() {
	s.stats.blocked = false
}

func (s *storeInfo) isBlocked() bool {
	return s.stats.blocked
}

func (s *storeInfo) isUp() bool {
	return s.GetState() == metapb.StoreState_Up
}

func (s *storeInfo) isOffline() bool {
	return s.GetState() == metapb.StoreState_Offline
}

func (s *storeInfo) isTombstone() bool {
	return s.GetState() == metapb.StoreState_Tombstone
}

func (s *storeInfo) downTime() time.Duration {
	return time.Since(s.stats.LastHeartbeatTS)
}

func (s *storeInfo) leaderCount() uint64 {
	return uint64(s.stats.LeaderCount)
}

func (s *storeInfo) leaderScore() float64 {
	return float64(s.stats.LeaderCount)
}

func (s *storeInfo) regionCount() uint64 {
	return uint64(s.stats.RegionCount)
}

func (s *storeInfo) regionScore() float64 {
	if s.stats.GetCapacity() == 0 {
		return 0
	}
	return float64(s.stats.RegionCount) / float64(s.stats.GetCapacity())
}

func (s *storeInfo) storageSize() uint64 {
	return s.stats.GetCapacity() - s.stats.GetAvailable()
}

func (s *storeInfo) storageRatio() float64 {
	if s.stats.GetCapacity() == 0 {
		return 0
	}
	return float64(s.storageSize()) / float64(s.stats.GetCapacity())
}

func (s *storeInfo) resourceCount(kind ResourceKind) uint64 {
	switch kind {
	case leaderKind:
		return s.leaderCount()
	case regionKind:
		return s.regionCount()
	default:
		return 0
	}
}

func (s *storeInfo) resourceScore(kind ResourceKind) float64 {
	switch kind {
	case leaderKind:
		return s.leaderScore()
	case regionKind:
		return s.regionScore()
	default:
		return 0
	}
}

func (s *storeInfo) getLabelValue(key string) string {
	for _, label := range s.GetLabels() {
		if label.GetKey() == key {
			return label.GetValue()
		}
	}
	return ""
}

func (s *storeInfo) getLocationID(keys []string) string {
	id := ""
	for _, k := range keys {
		v := s.getLabelValue(k)
		if len(v) == 0 {
			return ""
		}
		id += v
	}
	return id
}

// StoreStatus contains information about a store's status.
type StoreStatus struct {
	*pdpb.StoreStats

	// Blocked means that the store is blocked from balance.
	blocked         bool
	LeaderCount     uint32    `json:"leader_count"`
	LastHeartbeatTS time.Time `json:"last_heartbeat_ts"`
}

func newStoreStatus() *StoreStatus {
	return &StoreStatus{
		StoreStats: &pdpb.StoreStats{},
	}
}

func (s *StoreStatus) clone() *StoreStatus {
	return &StoreStatus{
		StoreStats:      proto.Clone(s.StoreStats).(*pdpb.StoreStats),
		blocked:         s.blocked,
		LeaderCount:     s.LeaderCount,
		LastHeartbeatTS: s.LastHeartbeatTS,
	}
}

// GetStartTS returns the start timestamp.
func (s *StoreStatus) GetStartTS() time.Time {
	return time.Unix(int64(s.GetStartTime()), 0)
}

// GetUptime returns the uptime.
func (s *StoreStatus) GetUptime() time.Duration {
	uptime := s.LastHeartbeatTS.Sub(s.GetStartTS())
	if uptime > 0 {
		return uptime
	}
	return 0
}
