export namespace backend {
	
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

export namespace hepsgasgun1 {
	
	export class Config {
	    InletAddr: number;
	    OutletAddr: number;
	    FireAddr: number;
	    VacuumRealseAddr: number;
	    PressureOpenAddr: number;
	    PressureCloseAddr: number;
	    TailVacuumPumpAddr: number;
	    TarVacuumPumpAddr: number;
	    InputPressureAddr: number;
	    VacuumFloatAddr: number;
	    PressureFloatAddr: number;
	    TailVaccumFloatAddr: number;
	    DogAddr: number;
	
	    static createFrom(source: any = {}) {
	        return new Config(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.InletAddr = source["InletAddr"];
	        this.OutletAddr = source["OutletAddr"];
	        this.FireAddr = source["FireAddr"];
	        this.VacuumRealseAddr = source["VacuumRealseAddr"];
	        this.PressureOpenAddr = source["PressureOpenAddr"];
	        this.PressureCloseAddr = source["PressureCloseAddr"];
	        this.TailVacuumPumpAddr = source["TailVacuumPumpAddr"];
	        this.TarVacuumPumpAddr = source["TarVacuumPumpAddr"];
	        this.InputPressureAddr = source["InputPressureAddr"];
	        this.VacuumFloatAddr = source["VacuumFloatAddr"];
	        this.PressureFloatAddr = source["PressureFloatAddr"];
	        this.TailVaccumFloatAddr = source["TailVaccumFloatAddr"];
	        this.DogAddr = source["DogAddr"];
	    }
	}

}

export namespace xinjiegasgun1 {
	
	export class Config {
	    InletAddr: number;
	    OutletAddr: number;
	    FireAddr: number;
	    VacuumRealseAddr: number;
	    PressureOpenAddr: number;
	    PressureCloseAddr: number;
	    TailVacuumPumpAddr: number;
	    TarVacuumPumpAddr: number;
	    VacuumFloatAddr: number;
	    PressureFloatAddr: number;
	    DogAddr: number;
	
	    static createFrom(source: any = {}) {
	        return new Config(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.InletAddr = source["InletAddr"];
	        this.OutletAddr = source["OutletAddr"];
	        this.FireAddr = source["FireAddr"];
	        this.VacuumRealseAddr = source["VacuumRealseAddr"];
	        this.PressureOpenAddr = source["PressureOpenAddr"];
	        this.PressureCloseAddr = source["PressureCloseAddr"];
	        this.TailVacuumPumpAddr = source["TailVacuumPumpAddr"];
	        this.TarVacuumPumpAddr = source["TarVacuumPumpAddr"];
	        this.VacuumFloatAddr = source["VacuumFloatAddr"];
	        this.PressureFloatAddr = source["PressureFloatAddr"];
	        this.DogAddr = source["DogAddr"];
	    }
	}

}

export namespace xinjiegasgun2 {
	
	export class Data {
	    InputPressure: number;
	    CylinderPressure: number;
	    PumpTubePressure: number;
	    PumpTubePressureHi: number;
	    TargetVacuumDegree: number;
	    TailVacuumDegree: number;
	
	    static createFrom(source: any = {}) {
	        return new Data(source);
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
	export class SwitchConfig {
	    Pressurize: number;
	    Decompress: number;
	    FireSwitch: number;
	    PumpTubePressurize: number;
	    PumpTubeDecompress: number;
	    PumpTubeVacuum: number;
	    TargetVacuum: number;
	    TailVacuumProtect: number;
	    PumpTubeProtect: number;
	    SystemDecompress: number;
	    TargetVacuumPump: number;
	    TailVacuumPump: number;
	
	    static createFrom(source: any = {}) {
	        return new SwitchConfig(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Pressurize = source["Pressurize"];
	        this.Decompress = source["Decompress"];
	        this.FireSwitch = source["FireSwitch"];
	        this.PumpTubePressurize = source["PumpTubePressurize"];
	        this.PumpTubeDecompress = source["PumpTubeDecompress"];
	        this.PumpTubeVacuum = source["PumpTubeVacuum"];
	        this.TargetVacuum = source["TargetVacuum"];
	        this.TailVacuumProtect = source["TailVacuumProtect"];
	        this.PumpTubeProtect = source["PumpTubeProtect"];
	        this.SystemDecompress = source["SystemDecompress"];
	        this.TargetVacuumPump = source["TargetVacuumPump"];
	        this.TailVacuumPump = source["TailVacuumPump"];
	    }
	}
	export class Config {
	    ip: string;
	    switches: SwitchConfig;
	    data: Data;
	
	    static createFrom(source: any = {}) {
	        return new Config(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.ip = source["ip"];
	        this.switches = this.convertValues(source["switches"], SwitchConfig);
	        this.data = this.convertValues(source["data"], Data);
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

