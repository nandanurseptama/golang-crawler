// Copyright The Golang Crawler Author
// SPDX-License-Identifier: Apache-2.0

// This file contains collection of youtube response struct
package youtube

type YtInitialDataResp struct {
	Contents struct {
		TwoColumn struct {
			PrimaryContents struct {
				SectionList struct {
					Contents []ContinuationItemResp `json:"contents"`
				} `json:"sectionListRenderer"`
			} `json:"primaryContents"`
		} `json:"twoColumnSearchResultsRenderer"`
	} `json:"contents"`
}

func (resp *YtInitialDataResp) GetVideoItems() []VideoItem {
	if resp == nil {
		return []VideoItem{}
	}
	var results []VideoItem
	for _, v := range resp.Contents.TwoColumn.PrimaryContents.SectionList.Contents {
		results = append(results, v.ItemSectionRenderer.GetVideoItems()...)
	}
	return results
}

type ContinuationItemResp struct {
	ItemSectionRenderer ItemSectionRendererResp `json:"itemSectionRenderer"`
}

type ItemSectionRendererResp struct {
	Contents []ContentResp `json:"contents"`
}

func (item *ItemSectionRendererResp) GetVideoItems() []VideoItem {
	if item == nil {
		return []VideoItem{}
	}

	if len(item.Contents) < 1 {
		return []VideoItem{}
	}

	var results []VideoItem
	for _, v := range item.Contents {
		if v.VideoRenderer.VideoID == "" {
			continue
		}

		results = append(results, v.VideoRenderer.ToVideoItem())
	}
	return results

}

type ContentResp struct {
	VideoRenderer VideoRendererResp `json:"videoRenderer"`
}

type VideoRendererResp struct {
	VideoID        string                         `json:"videoId"`
	Thumbnail      ThumbnailResp                  `json:"thumbnail"`
	Title          TitleResp                      `json:"title"`
	Length         SimpleTextResp                 `json:"lengthText"`
	ViewCount      SimpleTextResp                 `json:"viewCountText"`
	Owner          OwnerTextResp                  `json:"ownerText"`
	DetailMetadata []DetailedMetadataSnippetsResp `json:"detailedMetadataSnippets"`
	PublishedTime  SimpleTextResp                 `json:"publishedTimeText"`
}

func (v *VideoRendererResp) GetVideoDesc() string {
	if v == nil {
		return ""
	}

	if len(v.DetailMetadata) < 1 {
		return ""
	}

	return v.DetailMetadata[0].GetVideoDesc()
}

func (v *VideoRendererResp) ToVideoItem() VideoItem {
	if v == nil {
		return VideoItem{}
	}
	return VideoItem{
		ID:            v.VideoID,
		Channel:       v.Owner.GetChannel(),
		Thumbnails:    v.Thumbnail.Thumbnails,
		DurationText:  v.Length.GetText(),
		Duration:      parseDurationToSeconds(v.Length.GetText()),
		ViewCountText: v.ViewCount.GetText(),
		ViewCount:     parseViewCount(v.ViewCount.GetText()),
		Title:         v.Title.GetTitle(),
		Desc:          v.GetVideoDesc(),
		PublishedTime: v.PublishedTime.GetText(),
	}
}

type ThumbnailResp struct {
	Thumbnails []Thumbnail `json:"thumbnails"`
}

type TitleResp struct {
	Runs []struct {
		Text string `json:"text"`
	} `json:"runs"`
}

func (t *TitleResp) GetTitle() string {
	if t == nil {
		return ""
	}

	if len(t.Runs) < 1 {
		return ""
	}

	return t.Runs[0].Text
}

type SimpleTextResp struct {
	SimpleText string `json:"simpleText"`
}

func (t *SimpleTextResp) GetText() string {
	if t == nil {
		return ""
	}
	return t.SimpleText
}

type DetailedMetadataSnippetsResp struct {
	SnippetText struct {
		Runs []struct {
			Text string `json:"text"`
		} `json:"runs"`
	} `json:"snippetText"`
}

func (md *DetailedMetadataSnippetsResp) GetVideoDesc() string {
	if md == nil {
		return ""
	}

	if len(md.SnippetText.Runs) < 2 {
		return ""
	}

	return md.SnippetText.Runs[1].Text
}

type OwnerTextResp struct {
	Runs []struct {
		Text             string `json:"text"`
		NavigateEndpoint struct {
			BrowseEndpoint struct {
				BrowseId string `json:"browseId"`
				BaseUrl  string `json:"canonicalBaseUrl"`
			} `json:"browseEndpoint"`
		} `json:"navigationEndpoint"`
	} `json:"runs"`
}

func (owner *OwnerTextResp) GetChannel() Channel {
	if owner == nil {
		return Channel{}
	}

	if len(owner.Runs) < 1 {
		return Channel{}
	}
	first := owner.Runs[0]

	return Channel{
		Name:     first.Text,
		ID:       first.NavigateEndpoint.BrowseEndpoint.BrowseId,
		Endpoint: first.NavigateEndpoint.BrowseEndpoint.BaseUrl,
	}
}
