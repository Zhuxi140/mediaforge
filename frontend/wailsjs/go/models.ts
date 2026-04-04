export namespace media {
	
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

