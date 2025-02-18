export namespace backend {
	
	export class CaptureStats {
	    totalPackets: number;
	    tcpPackets: number;
	    udpPackets: number;
	    icmpPackets: number;
	    totalBytes: number;
	    startTime: number;
	
	    static createFrom(source: any = {}) {
	        return new CaptureStats(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.totalPackets = source["totalPackets"];
	        this.tcpPackets = source["tcpPackets"];
	        this.udpPackets = source["udpPackets"];
	        this.icmpPackets = source["icmpPackets"];
	        this.totalBytes = source["totalBytes"];
	        this.startTime = source["startTime"];
	    }
	}
	export class HTTPInfo {
	    method: string;
	    path: string;
	    version: string;
	    statusCode: number;
	    statusText: string;
	    contentType: string;
	    host: string;
	    isRequest: boolean;
	
	    static createFrom(source: any = {}) {
	        return new HTTPInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.method = source["method"];
	        this.path = source["path"];
	        this.version = source["version"];
	        this.statusCode = source["statusCode"];
	        this.statusText = source["statusText"];
	        this.contentType = source["contentType"];
	        this.host = source["host"];
	        this.isRequest = source["isRequest"];
	    }
	}
	export class Packet {
	    // Go type: time
	    timestamp: any;
	    protocol: string;
	    srcIP: string;
	    dstIP: string;
	    srcPort: number;
	    dstPort: number;
	    length: number;
	    info: string;
	    payload: string;
	    rawData: string;
	    httpInfo?: HTTPInfo;
	
	    static createFrom(source: any = {}) {
	        return new Packet(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.timestamp = this.convertValues(source["timestamp"], null);
	        this.protocol = source["protocol"];
	        this.srcIP = source["srcIP"];
	        this.dstIP = source["dstIP"];
	        this.srcPort = source["srcPort"];
	        this.dstPort = source["dstPort"];
	        this.length = source["length"];
	        this.info = source["info"];
	        this.payload = source["payload"];
	        this.rawData = source["rawData"];
	        this.httpInfo = this.convertValues(source["httpInfo"], HTTPInfo);
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
	export class PingResult {
	    // Go type: time
	    Timestamp: any;
	    RTT: number;
	    Success: boolean;
	    Error: string;
	    IP: string;
	
	    static createFrom(source: any = {}) {
	        return new PingResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Timestamp = this.convertValues(source["Timestamp"], null);
	        this.RTT = source["RTT"];
	        this.Success = source["Success"];
	        this.Error = source["Error"];
	        this.IP = source["IP"];
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
	export class TCPResult {
	    // Go type: time
	    Timestamp: any;
	    ConnectTime: number;
	    Success: boolean;
	    Error: string;
	    IP: string;
	
	    static createFrom(source: any = {}) {
	        return new TCPResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Timestamp = this.convertValues(source["Timestamp"], null);
	        this.ConnectTime = source["ConnectTime"];
	        this.Success = source["Success"];
	        this.Error = source["Error"];
	        this.IP = source["IP"];
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

