export namespace diagnostics {
	
	export class QuickCheckResult {
	    gatewayOk: boolean;
	    gatewayLatency: string;
	    internetOk: boolean;
	    internetLatency: string;
	    publicIp: string;
	
	    static createFrom(source: any = {}) {
	        return new QuickCheckResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.gatewayOk = source["gatewayOk"];
	        this.gatewayLatency = source["gatewayLatency"];
	        this.internetOk = source["internetOk"];
	        this.internetLatency = source["internetLatency"];
	        this.publicIp = source["publicIp"];
	    }
	}

}

export namespace network {
	
	export class Adapter {
	    name: string;
	    ipAddress: string;
	    enabled: boolean;
	    mode: string;
	
	    static createFrom(source: any = {}) {
	        return new Adapter(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.ipAddress = source["ipAddress"];
	        this.enabled = source["enabled"];
	        this.mode = source["mode"];
	    }
	}

}

export namespace profile {
	
	export class IPProfile {
	    id: string;
	    name: string;
	    ip: string;
	    subnet: string;
	    gateway: string;
	    dns: string;
	
	    static createFrom(source: any = {}) {
	        return new IPProfile(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.ip = source["ip"];
	        this.subnet = source["subnet"];
	        this.gateway = source["gateway"];
	        this.dns = source["dns"];
	    }
	}

}

