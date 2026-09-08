export namespace backend {
	
	export class APIResponse {
	    Status: boolean;
	    Message: string;
	
	    static createFrom(source: any = {}) {
	        return new APIResponse(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Status = source["Status"];
	        this.Message = source["Message"];
	    }
	}
	export class DataAddress {
	    InputPressure: number;
	    CylinderPressure: number;
	    PumpTubePressure: number;
	    PumpTubePressureHi: number;
	    TargetVacuumDegree: number;
	    TailVacuumDegree: number;
	
	    static createFrom(source: any = {}) {
	        return new DataAddress(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.InputPressure = source["InputPressure"];
	        this.CylinderPressure = source["CylinderPressure"];
	        this.PumpTubePressure = source["PumpTubePressure"];
	        this.PumpTubePressureHi = source["PumpTubePressureHi"];
	        this.TargetVacuumDegree = source["TargetVacuumDegree"];
	        this.TailVacuumDegree = source["TailVacuumDegree"];
	    }
	}
	export class SWITCHAddress {
	    Pressurize: number;
	    Decompress: number;
	    PumpTubePressurize: number;
	    PumpTubeDecompress: number;
	    PumpTubeVacuum: number;
	    TargetVacuum: number;
	    TailVacuumProtect: number;
	    PumpTubeProtect: number;
	    FireSwitch: number;
	    SystemDecompress: number;
	    TargetVacuumPump: number;
	    TailVacuumPump: number;
	
	    static createFrom(source: any = {}) {
	        return new SWITCHAddress(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Pressurize = source["Pressurize"];
	        this.Decompress = source["Decompress"];
	        this.PumpTubePressurize = source["PumpTubePressurize"];
	        this.PumpTubeDecompress = source["PumpTubeDecompress"];
	        this.PumpTubeVacuum = source["PumpTubeVacuum"];
	        this.TargetVacuum = source["TargetVacuum"];
	        this.TailVacuumProtect = source["TailVacuumProtect"];
	        this.PumpTubeProtect = source["PumpTubeProtect"];
	        this.FireSwitch = source["FireSwitch"];
	        this.SystemDecompress = source["SystemDecompress"];
	        this.TargetVacuumPump = source["TargetVacuumPump"];
	        this.TailVacuumPump = source["TailVacuumPump"];
	    }
	}
	export class GasGun2Config {
	    ip: string;
	    switches: SWITCHAddress;
	    dataAddresses: DataAddress;
	
	    static createFrom(source: any = {}) {
	        return new GasGun2Config(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.ip = source["ip"];
	        this.switches = this.convertValues(source["switches"], SWITCHAddress);
	        this.dataAddresses = this.convertValues(source["dataAddresses"], DataAddress);
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
	export class  {
	    name: string;
	    browser_download_url: string;
	
	    static createFrom(source: any = {}) {
	        return new (source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.browser_download_url = source["browser_download_url"];
	    }
	}
	export class GitHubRelease {
	    tag_name: string;
	    html_url: string;
	    assets: [];
	
	    static createFrom(source: any = {}) {
	        return new GitHubRelease(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.tag_name = source["tag_name"];
	        this.html_url = source["html_url"];
	        this.assets = this.convertValues(source["assets"], );
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

