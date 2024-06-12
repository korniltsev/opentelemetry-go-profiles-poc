package pprofconvert

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io"

	//todo do not use grafana/pyroscope api package
	googlepb "github.com/grafana/pyroscope/api/gen/proto/go/google/v1"

	otlppb "go.opentelemetry.io/proto/otlp/profiles/v1experimental"
)

func otlpProfileFromBytes(bs []byte) (*otlppb.Profile, error) {
	bs, err := decompress(bs)
	if err != nil {
		return nil, fmt.Errorf("failed to decompress pprof profile: %w", err)
	}
	googleProfile := new(googlepb.Profile)
	err = googleProfile.UnmarshalVT(bs)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal pprof profile: %w", err)
	}

	return otlpProfileFromGoogleProfile(googleProfile)
}

func decompress(bs []byte) ([]byte, error) {
	r, err := gzip.NewReader(bytes.NewReader(bs))
	if err != nil {
		return nil, fmt.Errorf("failed to create gzip reader: %w", err)
	}
	defer r.Close()
	return io.ReadAll(r)

}

func otlpProfileFromGoogleProfile(p *googlepb.Profile) (*otlppb.Profile, error) {

	mappings, mappingsMap := otlMappings(p.Mapping)
	functions, functionsMap := otlpFunctions(p.Function)
	locations, locationsMap := otlpLocations(p.Location, mappingsMap, functionsMap)
	samples, locationIndices := otlpSamples(p.Sample, locationsMap)
	res := &otlppb.Profile{
		SampleType:        otlpSampleTypes(p.SampleType),
		Sample:            samples,
		Mapping:           mappings,
		Location:          locations,
		LocationIndices:   locationIndices,
		Function:          functions,
		AttributeTable:    nil,
		AttributeUnits:    nil,
		LinkTable:         nil,
		StringTable:       p.StringTable,
		DropFrames:        p.DropFrames,
		KeepFrames:        p.KeepFrames,
		TimeNanos:         p.TimeNanos,
		DurationNanos:     p.DurationNanos,
		PeriodType:        otlpValueType(p.PeriodType),
		Period:            p.Period,
		Comment:           p.Comment,
		DefaultSampleType: p.DefaultSampleType,
	}
	return res, nil
}

func otlpSamples(sample []*googlepb.Sample, locationsId2Index map[uint64]int) ([]*otlppb.Sample, []int64) {
	res := make([]*otlppb.Sample, len(sample))
	locationIndices := make([]int64, 0)
	for sampleIndex, s := range sample {
		startIndex := len(locationIndices)
		for _, locID := range s.LocationId {
			locationIndices = append(locationIndices, int64(locationsId2Index[locID]))
		}
		endIndex := len(locationIndices)
		res[sampleIndex] = &otlppb.Sample{
			//LocationIndex:       nil,
			LocationsStartIndex: uint64(startIndex),
			LocationsLength:     uint64(endIndex - startIndex),
			StacktraceIdIndex:   0,
			Value:               s.Value,
			Label:               oltpLabels(s.Label),
			Attributes:          nil,
			Link:                0,
			TimestampsUnixNano:  nil,
		}
	}
	return res, locationIndices
}

func oltpLabels(label []*googlepb.Label) []*otlppb.Label {
	if label == nil {
		return nil
	}
	res := make([]*otlppb.Label, len(label))
	for i, l := range label {
		res[i] = &otlppb.Label{
			Key:     l.Key,
			Str:     l.Str,
			Num:     l.Num,
			NumUnit: l.NumUnit,
		}
	}
	return res

}

func otlpFunctions(function []*googlepb.Function) ([]*otlppb.Function, map[uint64]int) {
	id2index := make(map[uint64]int)
	res := make([]*otlppb.Function, len(function))
	for i, f := range function {
		res[i] = &otlppb.Function{
			//Id:              uint64(i),
			Name:       f.Name,
			SystemName: f.SystemName,
			Filename:   f.Filename,
			StartLine:  f.StartLine,
		}
		id2index[f.Id] = i
	}
	return res, id2index

}

func otlpLocations(location []*googlepb.Location, mappingsId2Index map[uint64]int, functionsId2Index map[uint64]int) ([]*otlppb.Location, map[uint64]int) {
	res := make([]*otlppb.Location, len(location))
	locationsMap := make(map[uint64]int)
	for i, l := range location {
		mappingIndex := mappingsId2Index[l.MappingId]
		res[i] = &otlppb.Location{
			//Id:              uint64(i),
			MappingIndex: uint64(mappingIndex),
			Address:      l.Address,
			Line:         otlpLines(l.Line, functionsId2Index),
			IsFolded:     l.IsFolded,
			TypeIndex:    0,
			Attributes:   nil,
		}
		locationsMap[l.Id] = i
	}
	return res, locationsMap
}

func otlpLines(line []*googlepb.Line, functionsId2Index map[uint64]int) []*otlppb.Line {

	res := make([]*otlppb.Line, len(line))
	for i, l := range line {
		res[i] = &otlppb.Line{
			FunctionIndex: uint64(functionsId2Index[l.FunctionId]),
			Line:          l.Line,
			Column:        0,
		}
	}
	return res
}

func otlMappings(mapping []*googlepb.Mapping) ([]*otlppb.Mapping, map[uint64]int) {
	id2index := make(map[uint64]int)
	res := make([]*otlppb.Mapping, len(mapping))
	for i, m := range mapping {
		res[i] = &otlppb.Mapping{
			//Id:              uint64(i),
			MemoryStart:     m.MemoryStart,
			MemoryLimit:     m.MemoryLimit,
			FileOffset:      m.FileOffset,
			Filename:        m.Filename,
			BuildId:         m.BuildId,
			BuildIdKind:     otlppb.BuildIdKind_BUILD_ID_LINKER,
			Attributes:      nil,
			HasFunctions:    m.HasFunctions,
			HasFilenames:    m.HasFilenames,
			HasLineNumbers:  m.HasLineNumbers,
			HasInlineFrames: m.HasInlineFrames,
		}
		id2index[m.Id] = i
	}
	return res, id2index
}

func otlpSampleTypes(sampleType []*googlepb.ValueType) []*otlppb.ValueType {
	res := make([]*otlppb.ValueType, len(sampleType))
	for i, st := range sampleType {
		res[i] = otlpValueType(st)
	}
	return res
}

func otlpValueType(st *googlepb.ValueType) *otlppb.ValueType {
	return &otlppb.ValueType{
		Type:                   st.Type,
		Unit:                   st.Unit,
		AggregationTemporality: otlppb.AggregationTemporality_AGGREGATION_TEMPORALITY_UNSPECIFIED,
	}
}
