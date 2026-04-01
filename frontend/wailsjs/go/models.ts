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
	    original_path?: string;
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
	        this.original_path = source["original_path"];
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

