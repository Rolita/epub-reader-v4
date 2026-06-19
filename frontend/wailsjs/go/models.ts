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

