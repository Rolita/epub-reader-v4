export namespace book {
	
	export class ImportResult {
	    success: boolean;
	    title: string;
	    author?: string;
	    coverUrl?: string;
	    md5: string;
	    filePath: string;
	    error?: string;
	
	    static createFrom(source: any = {}) {
	        return new ImportResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.success = source["success"];
	        this.title = source["title"];
	        this.author = source["author"];
	        this.coverUrl = source["coverUrl"];
	        this.md5 = source["md5"];
	        this.filePath = source["filePath"];
	        this.error = source["error"];
	    }
	}

}

export namespace main {
	
	export class FileInfo {
	    path: string;
	    size: number;
	    modTime: number;
	
	    static createFrom(source: any = {}) {
	        return new FileInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.path = source["path"];
	        this.size = source["size"];
	        this.modTime = source["modTime"];
	    }
	}

}

export namespace webdav {
	
	export class LogEntry {
	    timestamp: string;
	    action: string;
	    result: string;
	    type: string;
	
	    static createFrom(source: any = {}) {
	        return new LogEntry(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.timestamp = source["timestamp"];
	        this.action = source["action"];
	        this.result = source["result"];
	        this.type = source["type"];
	    }
	}

}

