export namespace config {
	
	export class Settings {
	    mediaOutputDir: string;
	    mediaHwAccel: string;
	    mediaQuality: string;
	    mediaVideoCRF: string;
	    mediaAudioVBR: string;
	    mediaTargetFmt: string;
	    subOutputDir: string;
	    subFormat: string;
	    subConvertDir: string;
	    subConvertFmt: string;
	
	    static createFrom(source: any = {}) {
	        return new Settings(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.mediaOutputDir = source["mediaOutputDir"];
	        this.mediaHwAccel = source["mediaHwAccel"];
	        this.mediaQuality = source["mediaQuality"];
	        this.mediaVideoCRF = source["mediaVideoCRF"];
	        this.mediaAudioVBR = source["mediaAudioVBR"];
	        this.mediaTargetFmt = source["mediaTargetFmt"];
	        this.subOutputDir = source["subOutputDir"];
	        this.subFormat = source["subFormat"];
	        this.subConvertDir = source["subConvertDir"];
	        this.subConvertFmt = source["subConvertFmt"];
	    }
	}

}

export namespace media {
	
	export class AudioStream {
	    index: number;
	    codec: string;
	    sampleRate: number;
	    channels: number;
	    bitRate: string;
	    language: string;
	
	    static createFrom(source: any = {}) {
	        return new AudioStream(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.index = source["index"];
	        this.codec = source["codec"];
	        this.sampleRate = source["sampleRate"];
	        this.channels = source["channels"];
	        this.bitRate = source["bitRate"];
	        this.language = source["language"];
	    }
	}
	export class FFmpegTask {
	    ID: string;
	    InputPath: string;
	    OutputPath: string;
	    Format: string;
	    Quality: string;
	    VideoCRF: string;
	    AudioVBR: string;
	    HwAccel: string;
	    ForceDropSubtitle: boolean;
	
	    static createFrom(source: any = {}) {
	        return new FFmpegTask(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.ID = source["ID"];
	        this.InputPath = source["InputPath"];
	        this.OutputPath = source["OutputPath"];
	        this.Format = source["Format"];
	        this.Quality = source["Quality"];
	        this.VideoCRF = source["VideoCRF"];
	        this.AudioVBR = source["AudioVBR"];
	        this.HwAccel = source["HwAccel"];
	        this.ForceDropSubtitle = source["ForceDropSubtitle"];
	    }
	}
	export class SubtitleStream {
	    Index: string;
	    Language: string;
	    Codec: string;
	
	    static createFrom(source: any = {}) {
	        return new SubtitleStream(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Index = source["Index"];
	        this.Language = source["Language"];
	        this.Codec = source["Codec"];
	    }
	}
	export class VideoStream {
	    index: number;
	    codec: string;
	    width: number;
	    height: number;
	    frameRate: string;
	    bitRate: string;
	    pixelFormat: string;
	    profile: string;
	
	    static createFrom(source: any = {}) {
	        return new VideoStream(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.index = source["index"];
	        this.codec = source["codec"];
	        this.width = source["width"];
	        this.height = source["height"];
	        this.frameRate = source["frameRate"];
	        this.bitRate = source["bitRate"];
	        this.pixelFormat = source["pixelFormat"];
	        this.profile = source["profile"];
	    }
	}
	export class MediaInfo {
	    filePath: string;
	    fileSize: string;
	    format: string;
	    duration: string;
	    bitRate: string;
	    video: VideoStream[];
	    audio: AudioStream[];
	    subtitle: SubtitleStream[];
	
	    static createFrom(source: any = {}) {
	        return new MediaInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.filePath = source["filePath"];
	        this.fileSize = source["fileSize"];
	        this.format = source["format"];
	        this.duration = source["duration"];
	        this.bitRate = source["bitRate"];
	        this.video = this.convertValues(source["video"], VideoStream);
	        this.audio = this.convertValues(source["audio"], AudioStream);
	        this.subtitle = this.convertValues(source["subtitle"], SubtitleStream);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	

}

export namespace renamer {
	
	export class ExtractRule {
	    Name: string;
	    Pattern: string;
	
	    static createFrom(source: any = {}) {
	        return new ExtractRule(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Name = source["Name"];
	        this.Pattern = source["Pattern"];
	    }
	}
	export class RenamePreview {
	    originalPath?: string;
	    originalName: string;
	    newName: string;
	    newPath: string;
	    hasConflict: boolean;
	    formatError: string;
	
	    static createFrom(source: any = {}) {
	        return new RenamePreview(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.originalPath = source["originalPath"];
	        this.originalName = source["originalName"];
	        this.newName = source["newName"];
	        this.newPath = source["newPath"];
	        this.hasConflict = source["hasConflict"];
	        this.formatError = source["formatError"];
	    }
	}
	export class RenameRule {
	    Mode: string;
	    Prefix: string;
	    Suffix: string;
	    ReplaceOld: string;
	    ReplaceNew: string;
	    SmartRules: ExtractRule[];
	    SmartTemplate: string;
	    CleanChars: string;
	
	    static createFrom(source: any = {}) {
	        return new RenameRule(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Mode = source["Mode"];
	        this.Prefix = source["Prefix"];
	        this.Suffix = source["Suffix"];
	        this.ReplaceOld = source["ReplaceOld"];
	        this.ReplaceNew = source["ReplaceNew"];
	        this.SmartRules = this.convertValues(source["SmartRules"], ExtractRule);
	        this.SmartTemplate = source["SmartTemplate"];
	        this.CleanChars = source["CleanChars"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}

}

