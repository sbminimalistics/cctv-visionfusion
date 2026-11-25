import "./App.css";
import { DashboardProvider } from "./DashboardState";
import { EdgeDeviceLiveStatus } from "./EdgeDeviceLiveStatus";

function App() {
    return (
        <>
            <DashboardProvider>
                <EdgeDeviceLiveStatus />
            </DashboardProvider>
        </>
    );
}

export default App;
